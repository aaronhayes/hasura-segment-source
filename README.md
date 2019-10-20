# Hasura Segment Source

[![Docker Automated build](https://img.shields.io/docker/automated/aaronhayes/hasura-segment-source?style=flat-square)](https://cloud.docker.com/repository/docker/aaronhayes/hasura-segment-source/general)
![GitHub](https://img.shields.io/github/license/aaronhayes/hasura-segment-source?style=flat-square)

> The easiest way to send Hasura Event Triggers to segment! Turbo charge your marketing, and analytics stacks!

## Contents
 - [Overview](#what-is-this)
 - [Configuration](#configuration)
 - [Deployment Guide](#deploying)
    - [Docker](#docker-compose)
    - [Zeit Now](#zeit-now)
    - [Segment Setup](#setting-up-segment)
    - [Hasura Setup](#setting-up-hasura)
 - [Roadmap](#roadmap)

## Overview

This is a lightweight Go server that connects [Hasura](https://hasura.io/) to [Segment](https://segment.com/) using Hasura's [Event Triggers](https://docs.hasura.io/1.0/graphql/manual/event-triggers/index.html) Feature. You can now easily your marketing and analytics stack using Segment whilst keeping your PostgreSQL DB as the single source of truth. The server exposes two endpoints on port `4004`.
1. `POST /webhook` - This is where Hasura needs to send the event triggers
2. `GET /health` - Healthcheck

## Configuration

Hasura Segment Source has a few configuration you need to setup.

| Environment Variable | Required | Default | Description |
| -------------------- | -------- | ------- | ----------- | 
| SEGMENT_WRITE_API_KEY | `true` | `null` | Your Segment Go Source write key.
| USER_ID_FIELD | `false` | `user_id` | The field container the user id. If this is incorrectly set or doesn't exist the user ID will be set to `anonymous`.

## Deploying
### Docker-compose

```yaml
version: '3.6'
services:
  # PostgreSQL database
  postgres:
    image: mdillon/postgis
    restart: always
    ports:
      - '5432:5432'
    volumes:
      - pgdata:/var/lib/postgresql/data
    logging:
      driver: 'json-file'
      options:
        max-size: '100k'
        max-file: '2'

  # Hasura Graphql Server
  hasura:
    image: hasura/graphql-engine:v1.0.0-beta.6
    ports:
      - '8080:8080'
    depends_on:
      - postgres
      - hasura-segment-source
    environment:
      HASURA_GRAPHQL_ENABLE_CONSOLE: 'true'
      HASURA_SEGMENT_SOURCE_WEBHOOK_URL: 'http://hasura-segment-source:4004/webhook'
    env_file:
      - hasura.env

  # Hasura Segment Source
  hasura-segment-source:
    image: aaronhayes/hasura-segment-source
    ports:
      - 4004:4004
    environemnt:
      - SEGMENT_WRITE_API_KEY: 'YOUR_SEGMENT_WRITE_KEY'
      - USER_ID_FIELD: 'user_id'

```

### Zeit Now

> Coming soon ...

### Setting Up Segment 

Follow this [Segment's Guide](https://segment.com/docs/guides/setup/how-do-i-find-my-write-key/) to help find your `SEGMENT_WRITE_API_KEY`. You will need to choose the `GoLang Server` Source.

### Setting Up Hasura

Follow Hasura's [Offical Guide](https://docs.hasura.io/1.0/graphql/manual/event-triggers/create-trigger.html) on creating an event trigger. Add the trigger to any insert/delete/update table you want to synced!

## Roadmap

- Zeit Now Config
- Actual versioning 
- Choose listening port
- Whitelist/Blacklist keys. This will allow you to control exactly which fields are sent to Segment
- Identify User Traits.  