#!/bin/sh
openssl req -x509 -newkey rsa:4096 -keyout az-dev-key.pem -out az-dev-cert.pem -sha256 -days 3650 -nodes -subj "/C=XX/ST=StateName/L=CityName/O=CompanyName/OU=CompanySectionName/CN=azurite" -addext "subjectAltName = DNS:azurite"
