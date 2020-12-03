Param(
	[Parameter(Mandatory = $false)]
	[int]
	$SleepInterval = 1
)

function New-AzServiceBusSasToken {  
	param( 
           
		[Parameter(Mandatory = $true)]
		[string]
		$Namespace,
		[Parameter(Mandatory = $true)]
		[string]
		$PolicyName,
		[Parameter(Mandatory = $true)]
		[string]
		$Key
	)

	$origin = [DateTime]"1/1/1970 00:00" 
	$Expiry = (Get-Date).AddDays(5)    

	#compute the token expiration time.
	$diff = New-TimeSpan -Start $origin -End $Expiry 
	$tokenExpirationTime = [Convert]::ToInt32($diff.TotalSeconds)

	#Create a new instance of the HMACSHA256 class and set the key to UTF8 for the size of $Key
	$hmacsha = New-Object -TypeName System.Security.Cryptography.HMACSHA256
	$hmacsha.Key = [Text.Encoding]::UTF8.GetBytes($Key)

	#create the string that will be used when cumputing the hash
	$stringToSign = [Web.HttpUtility]::UrlEncode($Namespace) + "`n" + $tokenExpirationTime

	#Compute hash from the HMACSHA256 instance we created above using the size of the UTF8 string above.
	$hash = $hmacsha.ComputeHash([Text.Encoding]::UTF8.GetBytes($stringToSign))
	#Convert the hash to base 64 string
	$signature = [Convert]::ToBase64String($hash)

	#create the token
	$token = [string]::Format([Globalization.CultureInfo]::InvariantCulture, `
			"SharedAccessSignature sr={0}&sig={1}&se={2}&skn={3}", `
			[Web.HttpUtility]::UrlEncode($Namespace), `
			[Web.HttpUtility]::UrlEncode($signature), `
			$tokenExpirationTime, $PolicyName) 
	return $token
}


function Send-AzServiceBusMessage {
	param(
		[Parameter(Mandatory = $true)]
		[string]
		$ResourceGroupName,
		[Parameter(Mandatory = $true)]
		[string]
		$NamespaceName,
		[Parameter(Mandatory = $true)]
		[string]
		$TopicName,    
		[Parameter(Mandatory = $false)]
		[string]
		$PolicyName = 'RootManageSharedAccessKey',
		[Parameter(Mandatory = $false)]
		[object]
		$Message,
		[Parameter(Mandatory = $false)]
		[string]
		$Token
	)

	#set up the parameters for the Invoke-WebRequest
	$headers = @{ "Authorization" = "$Token"; "Content-Type" = "application/json" }
	$uri = "https://$NamespaceName.servicebus.windows.net/$TopicName/messages"

	$resp = Invoke-WebRequest -Uri $uri -Headers $headers -Method Post -Body $($Message | ConvertTo-Json -Depth 10)
	"$($resp.StatusCode): $($resp.StatusDescription) RaceID: '$($Message.data.raceid)' Race: '$($Message.data.racename)'"
}


$ResourceGroup = 'dapr-bet'
$NameSpace = "dapr-bet-sbus"
$TopicName = "results"
$PolicyName = "RootManageSharedAccessKey"

$races = Get-Content ./results.json | ConvertFrom-Json

$serviceBussInstalled = Get-InstalledModule Az.ServiceBus
if (!$serviceBussInstalled) {
	Install-Module Az.ServiceBus
}

$ns = (Get-AzServiceBusNamespace -ResourceGroupName $ResourceGroup -Name $namespace).Name
$key = (Get-AzServiceBusKey -ResourceGroupName $ResourceGroup -Namespace $ns -Name $PolicyName).PrimaryKey  
$token = New-AzServiceBusSasToken -Namespace $ns -Policy $PolicyName -Key $key
$i = 0

1..100 | ForEach-Object {
	foreach ($race in $races) {
		Write-Host "sending message [$i]..." -ForegroundColor Yellow			
		Send-AzServiceBusMessage -ResourceGroupName $ResourceGroup `
			-NamespaceName $ns `
			-TopicName $TopicName `
			-PolicyName $PolicyName `
			-Message $race `
			-Token $token
		$i++
	}
}
