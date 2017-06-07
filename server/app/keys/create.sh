
## openssl genrsa -out app.rsa 1024
## openssl rsa -in app.rsa -pubout -outform PEM -out app.rsa.pub

ssh-keygen -t rsa -b 4096 -f jwtRS256.key
# Don't add passphrase
openssl rsa -in jwtRS256.key -pubout -outform PEM -out jwtRS256.key.pub
cat jwtRS256.key
cat jwtRS256.key.pub
