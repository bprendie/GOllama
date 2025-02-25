## GOllama

### This is a fairly simple text interface for chatting with your remote ollama 
server via the CLI.

#### Usage:
edit the llama.json file

{
    "ollamaURL": "http://your.ollama.server:11434/api/chat",
    "ollamaPort": 11434,
    "modelName": "llama3.2",
    "contextWindowSize": 100000,
    "systemPrompt": "you are a helpful, friendly assitant",
    "humanName": "Bob",
    "AIName": "Botty McGee"
  }
  

### compile and run:

go build llama.go

note: The ollama.json file needs to be in your working path in order for it to 
be picked up.

