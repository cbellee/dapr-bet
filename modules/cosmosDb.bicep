param location string
param tags object
param databases array
param prefix string

var cosmosAccountName = '${prefix}-cosmos-${uniqueString(resourceGroup().id)}'

resource cosmosAccount 'Microsoft.DocumentDB/databaseAccounts@2023-09-15' = {
  name: cosmosAccountName
  location: location
  tags: tags
  kind: 'GlobalDocumentDB'
  properties: {
    databaseAccountOfferType: 'Standard'
    locations: [
      {
        locationName: location
        failoverPriority: 0
      }
    ]
  }
}

resource cosmosDatabases 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases@2023-09-15' = [for db in databases: {
  name: db.name
  parent: cosmosAccount
  location: location
  tags: tags
  properties: {
    resource: {
      id: db.name
    }
    options: {
      autoscaleSettings: {
        maxThroughput: 1000
      }
    }
  }
}]

resource cosmosContainers 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers@2023-09-15' = [for (db, index) in databases: {
  name: 'default'
  parent: cosmosDatabases[index]
  properties: {
    options: {}
    resource: {
      id: 'default'
      partitionKey: {
        kind: 'Hash'
        paths: [
          db.partitionKey
        ]
      }
    }
  }
}]

output connectionString string = cosmosAccount.listConnectionStrings().connectionStrings[0].connectionString
output masterKey string = cosmosAccount.listKeys().primaryMasterKey
output url string = cosmosAccount.properties.documentEndpoint
