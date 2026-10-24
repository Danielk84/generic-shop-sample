> [!WARNING]
> This project created for fun and hobbies, and it does not have a lot of parts that needed for good product deployment.

# Generic-Shop
This is a simple multi seller shopping project based on Golang and Vue.

## Some Feature
- orders managing
- payment based on ZarinPal
- SQL query and migrations based on PostgreSQL
- background tasks for organizing Orders-Processing, unpaid Orders-expiration, Assigning the sold product to the seller based on Goroutins and Redis message queue
- Auth workflow managed by email temporary pass key validation
- There is a basic permission managing based on simple number enum logic that manage access
- storage managing based on AWS client lib and used seaweedfs for its backend
- client app used vue, tanStack-vue-query and axios for managing fronend logic and backend communication and tailwind for ui components
- Deploying backend and client with Nginx config
- Managing Cors and Security HTTP headers with Nginx
- Dockerise server and client
- Client data validator based on zod and server data validator based on validate v10 lib and golang tags
- Basic cli tools for server based on Cobra cli lib
- Organizing server Configurations with yaml
- Caching server Endpoins with Redis utils
- Using PlPgSQL for creating SQL utils and triggers for better data managing
- Search indexing based on full text search in PostgreSQL

## Run with docker compose
```bash
# Build containers
docker compose -f docker-compose.yaml build

# Start Running Services
docker compose -f docker-compose.yaml up
```
