#!/bin/bash

# 200 Success
echo '>>>>>>>> Get sync duties on valid slot'
curl -X 'GET' \
  'http://localhost:8080/api/v1/syncduties/11591367' \
  -H 'accept: application/json'

echo '>>>>>>>> Get block reward on valid slot'
curl -X 'GET' \
  'http://localhost:8080/api/v1/blockreward/11591367' \
  -H 'accept: application/json'

# 400 Bad Request
echo '>>>>>>>> Catch slot in the future error on sync duties'
curl -X 'GET' \
  'http://localhost:8080/api/v1/syncduties/99999999' \
  -H 'accept: application/json'

echo '>>>>>>>> Catch slot in the future error on block reward'
curl -X 'GET' \
  'http://localhost:8080/api/v1/blockreward/99999999' \
  -H 'accept: application/json'

# 404 Not Found
echo '>>>>>>>> Catch slot was missed on sync duties'
curl -X 'GET' \
  'http://localhost:8080/api/v1/syncduties/10000' \
  -H 'accept: application/json'
