#!/bin/bash
set -ex
cd ${BASH_SOURCE%/*}
curl -L http://localhost:5000/swagger/doc.json > ./config/csapi.json
yarn max openapi