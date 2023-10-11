param location string
param prefix string
param tags object

var kvName = '${prefix}-kv-${uniqueString(resourceGroup().id)}'

resource kv 'Microsoft.KeyVault/vaults@2023-02-01' = {
  name: kvName
  location: location
  tags: tags
  properties: {
    sku: {
      family: 'A'
      name: 'standard'
    }
    enableRbacAuthorization: true
    tenantId: tenant().tenantId
  }
}

output name string = kv.name
output id string = kv.id
