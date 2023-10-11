param location string
param prefix string
param retentionInDays int = 30

var name = '${prefix}-appinsights'

resource ai 'Microsoft.Insights/components@2020-02-02' = {
  name: name
  location: location
  kind: 'web'
  properties: {
    Application_Type: 'web'
    RetentionInDays: retentionInDays
  }
}

output instrumentationKey string = ai.properties.InstrumentationKey
output connectionString string = ai.properties.ConnectionString
