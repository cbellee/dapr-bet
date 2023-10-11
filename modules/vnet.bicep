param location string
param prefix string
param addressPrefix string
param subnets array
param tags object

var vnetName = '${prefix}-vnet'

resource vnet 'Microsoft.Network/virtualNetworks@2023-04-01' = {
  location: location
  tags: tags
  name: vnetName
  properties: {
    addressSpace: {
      addressPrefixes: [
        addressPrefix
      ]
    }
    subnets: [for subnet in subnets: {
      name: subnet.name
      properties: {
        addressPrefix: subnet.addressPrefix
      }
    }]
  }
}

output subnets array = vnet.properties.subnets
