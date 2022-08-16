# Redis-based QPS and total query limiter

This is an implementation of query/sec and total query limiter using Redis. We are making use of [SET with expire](https://redis.io/commands/set/), [INCR](https://redis.io/commands/incr/) and [GET](https://redis.io/commands/get/) to implement the limiter.

Due to the [EXPIRE passive behaviour](https://redis.io/commands/expire/), we were observing that redis will either miss deleting the key when TTL has past or delete with some delay. This was giving us false "Too many requests" especially in the case of QPS.
So, we also added TTL check for the key using [PTTL](https://redis.io/commands/pttl/) and deleting it using [DEL](https://redis.io/commands/del/), if the TTL is approaching its expiry.



## How to work with this repo

1. Set the following env variables in a `.env` file in the project root directory:
```bash
API_PORT=3000
API_MODE=debug
REDIS_ADDR=redis:6379
QPS_LIMIT=10
QUERY_LIMIT=100
```

1. Build the application using `make app`

1. Bring up containers using `docker-compose up`

1. Overwrite the qps and query limit constants in `test/main.go` file to what you have in `.env` file. The variables in `test/main.go` must correspond to the values in `.env` file.

1. Run test using `go run test/main.go`. Your will see the result logs in the [results](results/) folder. There are some logs already for the following inputs:
    - Query limit: 10000, QPS limit: 5
    - Query limit: 10000, QPS limit: 10
    - Query limit: 10000, QPS limit: 50

