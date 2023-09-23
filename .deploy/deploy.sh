#!/bin/sh
docker compose pull service
docker stop service
docker compose up -d