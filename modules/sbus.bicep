param location string
param tags object
param prefix string

var sbusName = '${prefix}-sbus-${uniqueString(resourceGroup().id)}'
var serviceBusEndpoint = '${serviceBus.id}/AuthorizationRules/RootManageSharedAccessKey'

resource serviceBus 'Microsoft.ServiceBus/namespaces@2022-10-01-preview' = {
  name: sbusName
  tags: tags
  location: location
  sku: {
    name: 'Standard'
  }
}

output name string = serviceBus.name
output id string = serviceBus.id
output apiVersion string = serviceBus.apiVersion
output connectionString string = listKeys(serviceBusEndpoint, serviceBus.apiVersion).primaryConnectionString
