# Configuration file

The configuration file needs to be a `json` file. The viper parser supports other formats but they haven't been as tested as the recommended one.

The easiest way to create or edit a configuration file is using the [KrakenDesigner](http://www.krakend.io/designer/)

## Json example


    {
	"version": 1,
	"name": "My lovely gateway",
	"port": 8080,
	"timeout": 10,
	"cache_ttl": 3600,
	"host": [
		"http://127.0.0.1:8080",
		"http://127.0.0.2:8000",
		"http://127.0.0.3:9000",
		"http://127.0.0.4"
	],
	"endpoints": [{
			"endpoint": "/users/{user}",
			"method": "GET",
			"backend": [{
					"host": [
						"http://127.0.0.3:9000",
						"http://127.0.0.4"
					],
					"url_pattern": "/registered/{user}",
					"whitelist": [
						"some",
						"what"
					],
					"mapping": {
						"email": "personal_email"
					}
				},
				{
					"host": [
						"http://127.0.0.1:8080"
					],
					"url_pattern": "/users/{user}/permissions",
					"blacklist": [
						"spam2",
						"notwanted2"
					]
				}
			],
			"concurrent_calls": 2,
			"timeout": 1000,
			"cache_ttl": 3600,
			"querystring_params": [
				"page",
				"limit"
			]
		},
		{
			"endpoint": "/foo/bar",
			"method": "POST",
			"backend": [{
				"host": [
					"https://127.0.0.1:8081"
				],
				"url_pattern": "/__debug/tupu"
			}],
			"concurrent_calls": 1,
			"timeout": 10000,
			"cache_ttl": 3600
		},
		{
			"endpoint": "/github",
			"method": "GET",
			"backend": [{
				"host": [
					"https://api.github.com"
				],
				"url_pattern": "/",
				"whitelist": [
					"authorizations_url",
					"code_search_url"
				]
			}],
			"concurrent_calls": 2,
			"timeout": 1000,
			"cache_ttl": 3600
		},
		{
			"endpoint": "/combination/{id}/{supu}",
			"method": "GET",
			"backend": [{
					"group": "first_post",
					"host": [
						"https://jsonplaceholder.typicode.com"
					],
					"url_pattern": "/posts/{id}?supu={supu}",
					"blacklist": [
						"userId"
					]
				},
				{
					"host": [
						"https://jsonplaceholder.typicode.com"
					],
					"url_pattern": "/users/{id}",
					"mapping": {
						"email": "personal_email"
					}
				}
			],
			"concurrent_calls": 3,
			"timeout": 4000,
			"querystring_params": [
				"page",
				"limit"
			]
		}
	]
}
