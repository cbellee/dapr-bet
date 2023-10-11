param topics array
param sbusNamespaceName string

resource sbusNamespace 'Microsoft.ServiceBus/namespaces@2022-10-01-preview' existing = {
  name: sbusNamespaceName
}

resource sbusTopics 'Microsoft.ServiceBus/namespaces/topics@2022-10-01-preview' = [for topic in topics: {
  name: topic
  parent: sbusNamespace
  properties: {}
}]

resource sbusSubscriptions 'Microsoft.ServiceBus/namespaces/topics/subscriptions@2022-10-01-preview' = [for (topic, index) in topics: {
  name: topic
  parent: sbusTopics[index]
  dependsOn: [
    sbusTopics[index]
  ]
}]
