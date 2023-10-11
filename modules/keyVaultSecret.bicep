param kvName string
param secretName string

@secure()
param secret string

resource kv 'Microsoft.KeyVault/vaults@2023-02-01' existing = {
  name: kvName
}

resource keyVaultSecret 'Microsoft.KeyVault/vaults/secrets@2021-04-01-preview' = {
  parent: kv
  name: secretName
  properties: {
    value: secret
  }
}

output secretUri string = keyVaultSecret.properties.secretUri
