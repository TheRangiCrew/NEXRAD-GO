module github.com/TheRangiCrew/NEXRAD-GO/server

go 1.22.1

replace github.com/TheRangiCrew/NEXRAD-GO/level2/ => ../level2/

replace github.com/TheRangiCrew/NEXRAD-GO/level2/nexrad => ../level2/nexrad

require github.com/aws/aws-sdk-go-v2 v1.26.1

require (
	github.com/TheRangiCrew/NEXRAD-GO/level2 v0.0.0-20240419005628-d98dda7ac56e // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.11 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.23.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.6 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/paulmach/go.geojson v1.5.0 // indirect
	github.com/surrealdb/surrealdb.go v0.2.1 // indirect
)

require (
	github.com/TheRangiCrew/NEXRAD-GO/level2/nexrad v0.0.0-20240422075631-1de2515c31c9
	github.com/aws/aws-sdk-go-v2/config v1.27.11
	github.com/aws/aws-sdk-go-v2/service/s3 v1.53.1
	github.com/aws/aws-sdk-go-v2/service/sqs v1.31.4
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/joho/godotenv v1.5.1
)
