param location string
param prefix string
param tags object
param retentionInDays int = 30

@allowed([
  'PerGB2018'
])
param sku string = 'PerGB2018'

var workspaceName = '${prefix}-${uniqueString(resourceGroup().id)}-wks'

resource azureMonitorWorkspace 'Microsoft.OperationalInsights/workspaces@2022-10-01' = {
  location: location
  name: workspaceName
  tags: tags
  properties: {
    retentionInDays: retentionInDays
    sku: {
      name: sku
    }
  }
}

output workspaceId string = azureMonitorWorkspace.id 
