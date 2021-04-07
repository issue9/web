#!/bin/bash

# SPDX-License-Identifier: MIT
 
openssl req -newkey rsa:2048 -x509 -nodes -keyout key.pem -new -out cert.pem -config req.cnf -sha256 -days 3650
