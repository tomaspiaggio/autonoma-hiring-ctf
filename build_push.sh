#!/bin/bash

az login                                                
az acr login --name autonomactfregistry

docker build -t autonomactfregistry.azurecr.io/ctf/ssh:1.0.14 --platform linux/amd64 --push .