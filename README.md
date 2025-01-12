![Build](https://github.com/NLipatov/goproxy/actions/workflows/main.yml/badge.svg)
[![codecov](https://codecov.io/gh/NLipatov/goproxy/branch/main/graph/badge.svg)](https://codecov.io/gh/NLipatov/goproxy)
![License](https://img.shields.io/badge/license-MIT-blue.svg?style=plastic)
![Stars](https://img.shields.io/github/stars/NLipatov/goproxy.svg)
![Forks](https://img.shields.io/github/forks/NLipatov/goproxy.svg)
![Issues](https://img.shields.io/github/issues/NLipatov/goproxy.svg)

goproxy is an HTTP(S)-proxy server.

It has a microservice architecture, with each microservice running in a separate container.
Microservices use a Kafka bus to transmit events to one another.

# Services

Check the docker-compose.yml file in the project's root directory

## proxy
The core service, used as an HTTP-proxy server.

The proxy uses an auth database to authorize clients to access the proxy service.
Only existing users can use the proxy.

### Domain events
Consumes:
1) `UserExceededTrafficLimitEvent` - triggers user restrictions;
2) `UserConsumedTrafficWithoutPlan` - triggers user restrictions.

Produces:
1) `UserConsumedTrafficEvent`

## rest-api 
Used to get, create, update, and delete users

## plan-controller
Consumes `UserConsumedTrafficEvent`, which is produced by proxy when a user has used it.

### Domain events
Consumes:
1) `UserConsumedTrafficEvent` - triggers updates to user traffic stats and validates user plans. 

Produces:
1) `UserExceededTrafficLimitEvent` - triggers user restrictions;
2) `UserConsumedTrafficWithoutPlan` - triggers user restrictions.

# Databases

There are 3 separate databases:
1) postgres - used for user authentication;
2) plan-postgres - used to store plans and user-plan affiliations.

## Migrations

The migration process is handled by migrator.go, located in the Data Access Layer (DAL).
All migration files are placed in directories with the same name as the corresponding database. 

For example:
- Proxy database migrations are stored in the proxydb directory;
- Plan database migrations are stored in the plans directory.


