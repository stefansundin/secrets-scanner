# secrets-scanner

Find API tokens and other secrets in your code.

Example: Search Git history including commit messages:
```
$ go build
$ git log -p | ./secrets-scanner
Found Redis URLs:
- 'redis://h:1i0647a29e4qsp3iefbttnhnca3@example.com:11141'
```

Use `-test` to test the found keys automatically. Not supported for all services.
```
$ git log -p | ./secrets-scanner -test
Found Redis URLs:
- 'redis://h:1i0647a29e4qsp3iefbttnhnca3@example.com:11141'
Connection failed: dial tcp 93.184.216.34:11141: i/o timeout
```

# Supported

- AWS Access Key ID
- Google Access Token
- Google API key
- Slack `xoxp` and `xoxb` tokens
- Redis URLs
- Gemfury URLs
