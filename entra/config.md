## Microservices
### Mocking Monitor Backend
* Exposes read and write api, admin and call-center roles will alter what the user is allwed to query for
* It has redirect url that shows that users are allowed to access the app via authorization code flow 
### Customer Service
Exposes read and write api
### Healer Service
Exposes read and write api
### Ladder Service
Exposes read and write api

### Clients
* Mocking Monitor Backend when acting as client - will consume 'HealerService', 'CustomerService' and 'LadderService' api
* When asking for access token - will use the .default scope for each of the target services. Via it's group memebership it will get read role for HealerService and CustomerService. For LadderService it will get read and write roles.
* Mocking Monitor Frontend - will be used by frontend app. Will use the .default scope for Mocking Monitor Backend when asking for access token. Will get scopes according to the group memberships for the end user.

### Users
Users will get the following roles for Mocking Monitor Backend when logging on, frontend will use Mocking Monitor Frontend client and ask for .default scope for Mocking Monitor Backend. 
- healer, will get read and write role 
- diego, will get admin role
- anyuser, will get no role
- callcenter, will call-center role

### Verifications

During the verification only password and client-secret provided in config should work
Audience shuold be handled correctly as well

Using the Mocking Monitor Backend client will get correct roles when asking for access token with different scopes using client id and and client secret
For instance asking for default scopes for "api://bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb" (CustomerService) will give it read role
- HealerService - read role
- LadderService - read and write role
- FuzzService - no roles
When trying to use the client id for auth code flow it should not be permitted giving the response that entra would give

Using the Mocking Monitor Frontent client will get correct roles when asking for access token with auth code flow and real users
with scope for default scope and Mocking Monitor Backend, "api://aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa" 
- healer, will get read and write role
- diego, will get admin role
- anyuser, will get no role
- callcenter, will call-center role

When trying to use default scope for HealerService it it should not be permitted giving the response that entra would give
When trying to use the the client for Credentials flow it should be permitted giving he response entra would give