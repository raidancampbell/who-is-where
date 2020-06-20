# who-is-where

## What

server tracking via easy HTTP GET requests

## Why

Dynamic DNS is sometimes too annoying to set up, or I don't want secrets on the host I'm tracking.

## How

clients send HTTP GET requests to the server endpoint `/api/v1/:zone/:host`. 
The server reads the client's address from the HTTP request, and stores it along with the given zone and host.
To retrieve the data, an HTTP GET request to `/api/v1/:zone` will dump all hosts and their addresses.
