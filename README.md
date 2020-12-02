# azure-golang
## Steps to set up the environment
Creating Service Principal using [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) 
Create the service principal with certificate
```
$: az ad sp create-for-rbac --name application --create-cert --sdk-auth
```
output
```
{
    "appId": <client id>,
    "displayName": "application",
    "fileWithCertAndPrivateKey": <the certificate file created on you machine>,
    "name": "http://application",
    "password": null,
    "tenant": <tenant id>
}
```
Now you need to convert the certificate to pkcs12
```
$: openssl pkcs12 -export -in certificate.pem -out certificate.pkcs12
```
	
Export the env variables
```
export AZURE_CLIENT_ID=<client id>
export AZURE_CERTIFICATE_PATH=<pkcs12 certificate>
export AZURE_TENANT_ID=<tenand id>
export AZURE_SUBSCRITPION=<subscription>
```
