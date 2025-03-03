# AOG

## Quick Starter Guide

### Prerequisites

AOG bridges your application to actual AI Services. You may need to host AI
services locally on the PC.

As a prototype, you may install ollama first and setup a local
AI service to be connected into AOG.

### Build AOG

You need to have `golang` installed first. And then, and then run command like this on Windows

```sh
go build -o aog.exe
```



### Configure AOG

- Rename `aog.config.example.json` to `aog.json`
- Edit `aog.json` by changing related configurations inside it.
  - In particular, the URLs of remote and local service providers

### Start the AOG Gateway on PC

Typically, you can simply run `aog` in your command line. It will start a
gateway running at `http://localhost:16688`

You may type `aog -help` to see all of the options.

- `-config` to change the path to AOG configuration
- `-port PORT` to change the port. Default is `16688`
- `-v` and `-vv` to show more or even more verbose information
- `-logHTTP filepath` will log the HTTP messages through AOG. Only for Debug
  purpose

### Update your Application to use AOG

Change your current application. For example, previously your application maybe
visit the ollama Restful AI service through `http://ollama/api/chat`. Now, you
may change it to access corresponding AOG service, e.g.
`http://localhost:16688/aog/api_flavors/ollama/api/chat`

### Examples of Running AOG

After AOG is built, you may simply run this on Windows

```sh
aog.exe -config your_aog_config.json
```

If you want to see some details for debugging purpose, you may try the following
which shows more verbose message and logs HTTP messages

```sh
aog.exe -v -config your_aog_config.json -log_http http.log
```

You may use `curl` to try it. These are examples on Windows. You may change
`stream` to `true` for stream mode.

```sh

# test chat in standard AOG style
curl http://localhost:16688/aog/v0.1/aog/services/chat -X POST -H "Content-Type: application/json" -d "{\"model\":\"llama3.1\",\"messages\":[{\"role\":\"user\",\"content\":\"why is the sky blue?\"}],\"options\":{\"seed\":12345,\"temperature\":0},\"stream\":false}"

# test generate in standard AOG style
curl http://localhost:16688/aog/v0.1/aog/services/generate -X POST -H "Content-Type: application/json" -d "{\"model\":\"llama3.1\",\"prompt\":\"whyistheskyblue?\",\"options\":{\"seed\":12345,\"temperature\":0},\"stream\":false}"

# test embed in standard AOG style
curl http://localhost:16688/aog/v0.1/aog/services/embed -X POST -H "Content-Type: application/json" -d "{\"model\":\"all-minilm\",\"input\":\"whyistheskyblue?\",\"options\":{\"seed\":12345,\"temperature\":0},\"stream\":false}"


# test ollama flavor - chat
curl http://localhost:16688/aog/v0.1/aog/api_flavors/ollama/api/chat -X POST -H "Content-Type: application/json" -d "{\"model\":\"llama3.1\",\"messages\":[{\"role\":\"user\",\"content\":\"why is the sky blue?\"}],\"options\":{\"seed\":12345,\"temperature\":0},\"stream\":false}"

# test ollama flavor - generate
curl http://localhost:16688/aog/v0.1/aog/api_flavors/ollama/api/generate -X POST -H "Content-Type: application/json" -d "{\"model\":\"llama3.1\",\"prompt\":\"whyistheskyblue?\",\"options\":{\"seed\":12345,\"temperature\":0},\"stream\":false}"

# test ollama flavor - embed
curl http://localhost:16688/aog/v0.1/aog/api_flavors/ollama/api/embed -X POST -H "Content-Type: application/json" -d "{\"model\":\"all-minilm\",\"input\":\"whyistheskyblue?\",\"options\":{\"seed\":12345,\"temperature\":0},\"stream\":false}"


# test openai flavor - chat
curl http://localhost:16688/aog/v0.1/aog/api_flavors/openai/v1/chat/completions -X POST -H "Content-Type: application/json"  -d "{\"model\":\"lmstudio-community/Meta-Llama-3.1-8B-Instruct-GGUF\",\"messages\":[{\"role\":\"system\",\"content\":\"Always answer in rhymes.\"},{\"role\":\"user\",\"content\":\"Introduce yourself.\"}],\"temperature\":0.7,\"max_tokens\":-1,\"stream\":false}"

```
