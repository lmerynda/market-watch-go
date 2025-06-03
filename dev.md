# Quadrant vector db

vscode ➜ /workspaces/market-watch-go (master) $ docker run -d   --name qdrant   -p 6333:6333   qdrant/qdrant:latest
docker: Error response from daemon: Conflict. The container name "/qdrant" is already in use by container "7312ff3871c38d6c95842b93efc7d5ee96b4055627be467be58041c65c30b4aa". You have to remove (or rename) that container to be able to reuse that name.

Run 'docker run --help' for more information
vscode ➜ /workspaces/market-watch-go (master) $ docker docker ps -a --filter "name=qdrant"^C
vscode ➜ /workspaces/market-watch-go (master) $ docker ps -a --filter "name=qdrant"
CONTAINER ID   IMAGE                  COMMAND             CREATED      STATUS                      PORTS                                                   NAMES
7312ff3871c3   qdrant/qdrant:latest   "./entrypoint.sh"   2 days ago   Exited (255) 14 hours ago   0.0.0.0:6333->6333/tcp, [::]:6333->6333/tcp, 6334/tcp   qdrant
vscode ➜ /workspaces/market-watch-go (master) $ docker start qdrant
qdrant


# Ollama based embeddings

http://host.docker.internal:11434
curl http://127.0.0.1:11434/v1/embeddings   -H "Content-Type: application/json"   -d '{
    "model": "nomic-embed-text:latest",
    "input": "hello world"
  }'
