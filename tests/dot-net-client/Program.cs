using System;
using System.Threading.Tasks;
using System.Collections.Generic;
using System.Net.Http;
using System.Linq;
using System.Threading;
using Azure.Messaging.ServiceBus;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Configuration;
using Microsoft.Identity.Web;

namespace dot_net_client
{
    class Program
    {
        static async Task<int> Main(string[] args)
        {
            if (args.Length < 1)
            {
                PrintUsage();
                return 1;
            }

            return args[0].ToLower() switch
            {
                "sb" => await RunServiceBusAsync(args),
                "identity" => await RunIdentityTestAsync(args),
                _ => await RunServiceBusLegacyAsync(args) // Support old format
            };
        }

        static void PrintUsage()
        {
            Console.WriteLine("Usage:");
            Console.WriteLine("  dot-net-client sb <connection-string> <topic-name> [subscription-name]");
            Console.WriteLine("  dot-net-client identity [wiremock-url]");
        }

        static async Task<int> RunServiceBusAsync(string[] args)
        {
            if (args.Length < 3)
            {
                Console.WriteLine("Error: Connection string and topic name are required for 'sb' command.");
                return 1;
            }
            return await ExecuteServiceBusLogic(args[1], args[2], args.Length > 3 ? args[3] : "test-subscription");
        }

        static async Task<int> RunServiceBusLegacyAsync(string[] args)
        {
            if (args.Length < 2) return 1;
            return await ExecuteServiceBusLogic(args[0], args[1], args.Length > 2 ? args[2] : "test-subscription");
        }

        static async Task<int> ExecuteServiceBusLogic(string connectionString, string topicName, string subscriptionName)
        {
            try
            {
                Console.WriteLine($"[DEBUG_LOG] Initializing Service Bus test: topic={topicName}, subscription={subscriptionName}");
                using var cts = new CancellationTokenSource(TimeSpan.FromSeconds(5)); // Increased to 5s for reliability in CI
                
                await using var client = new ServiceBusClient(connectionString, new ServiceBusClientOptions
                {
                    RetryOptions = new ServiceBusRetryOptions { TryTimeout = TimeSpan.FromSeconds(1) }
                });
                
                Console.WriteLine("[DEBUG_LOG] Creating sender...");
                var sender = client.CreateSender(topicName);
                
                Console.WriteLine("[DEBUG_LOG] Sending message...");
                await sender.SendMessageAsync(new ServiceBusMessage($"Hello Service Bus! Sent at {DateTime.Now}"), cts.Token);
                Console.WriteLine("Sent message successfully.");

                Console.WriteLine("[DEBUG_LOG] Creating receiver...");
                var receiver = client.CreateReceiver(topicName, subscriptionName);
                
                Console.WriteLine("[DEBUG_LOG] Receiving message (max 5s wait, but 5s client-side cancellation)...");
                var receivedMessage = await receiver.ReceiveMessageAsync(TimeSpan.FromSeconds(5), cts.Token);
                
                if (receivedMessage != null)
                {
                    Console.WriteLine($"Received message: {receivedMessage.Body}");
                    Console.WriteLine("[DEBUG_LOG] Completing message...");
                    await receiver.CompleteMessageAsync(receivedMessage, cts.Token);
                    Console.WriteLine("[DEBUG_LOG] Successfully finished!");
                    return 0;
                }
                
                Console.WriteLine("[DEBUG_LOG] No message received.");
                return 1;
            }
            catch (OperationCanceledException)
            {
                Console.WriteLine("ERROR: Operation timed out after 1 second.");
                return 1;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"ERROR: {ex.Message}");
                if (ex.InnerException != null) 
                {
                    Console.WriteLine($"INNER ERROR: {ex.InnerException.Message}");
                }
                return 1;
            }
        }

        static async Task<int> RunIdentityTestAsync(string[] args)
        {
            string wiremockUrl = args.Length > 1 ? args[1] : "https://wiremock.local";
            try
            {
                Console.WriteLine("Setting up WebClient with Microsoft Identity defaults...");

                var configuration = new ConfigurationBuilder()
                    .AddInMemoryCollection(new Dictionary<string, string>
                    {
                        // Minimum viable Identity dictionary overriding zeroes
                        {"AzureAd:Instance", "https://login.microsoftonline.com/"},
                        {"AzureAd:TenantId", "b5a920d6-7d3c-44fe-baad-4ffed6b8774d"},
                        {"AzureAd:ClientId", "e697b97c-9b4b-487f-9f7a-248386f78864"}, // Active identity claims payload
                        {"AzureAd:ClientSecret", "dummy-secret"}
                    }!)
                    .Build();

                var services = new ServiceCollection();
                services.AddSingleton<IConfiguration>(configuration);
                
                // Keep chaining entirely to defaults setup relying purely on Configuration binding structure
                services.AddMicrosoftIdentityWebAppAuthentication(configuration)
                    .EnableTokenAcquisitionToCallDownstreamApi()
                    .AddInMemoryTokenCaches();

                services.AddHttpClient("wiremock", client =>
                {
                    client.BaseAddress = new Uri(wiremockUrl);
                    client.Timeout = TimeSpan.FromSeconds(1);
                });

                var serviceProvider = services.BuildServiceProvider();

                // Requirement Verification: Ensure client actively connects & acquires identity accessToken
                Console.WriteLine("Acquiring token instance verifying token identity...");
                try 
                {
                    using var cts = new CancellationTokenSource(TimeSpan.FromSeconds(1));
                    var tokenAcquisition = serviceProvider.GetRequiredService<ITokenAcquisition>();
                    // Some environments might need a real-looking scope
                    var token = await tokenAcquisition.GetAccessTokenForAppAsync("api://e697b97c-9b4b-487f-9f7a-248386f78864/.default");
                    
                    if (!string.IsNullOrEmpty(token))
                    {
                        Console.WriteLine("Successfully acquired access token!");
                    }
                    else
                    {
                        Console.WriteLine("Warning: Handshake processed but identity token value returned explicitly empty.");
                        return 1;
                    }
                }
                catch (Exception ex)
                {
                     Console.WriteLine($"Token verification error encountered mapping OIDC validation constraints: {ex.Message}");
                     return 1;
                }

                // Verify HTTPS client connection behavior into remote proxy node limits
                var httpClientFactory = serviceProvider.GetRequiredService<IHttpClientFactory>();
                var httpClient = httpClientFactory.CreateClient("wiremock");
                
                Console.WriteLine($"Attempting downstream verification Wiremock connection ({wiremockUrl})...");
                var response = await httpClient.GetAsync("/__admin/mappings");

                Console.WriteLine($"Response network HTTP status footprint matched: {response.StatusCode}");
                if (response.IsSuccessStatusCode)
                {
                    Console.WriteLine("Successfully hit wiremock!");
                    return 0;
                }
                
                return 1;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"FATAL IDENTITY BOOT FAILURE: {ex.Message}");
                if (ex.InnerException != null) 
                {
                    Console.WriteLine($"[Inner Exception Trigger] {ex.InnerException.Message}");
                }
                return 1;
            }
        }
    }
}
