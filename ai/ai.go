package ai

import (
	"context"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"log"
)

func InitVectorStore() {
	// Initialize Pinecone vector store
	client, err := pinecone.New("your-pinecone-api-key", "your-pinecone-environment")
	if err != nil {
		return nil, err
	}

	return client, nil
}

func AskQuestion(question string) (string, error) {
	InitVectorStore()
	llm, err := ollama.New(ollama.WithModel("llama3.2"))

	if err != nil {
		log.Fatal(err)
	}

	// Create an embedding model (e.g., OpenAI or Ollama embeddings)

	query := question

	ctx := context.Background()
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, query)
	if err != nil {
		log.Fatal(err)
	}

	return completion, nil
}

func AddDocument(client *pinecone.Pinecone, id string, content string) error {
	// Initialize the embedding model
	embeddingModel, err := embeddings.NewOpenAIEmbedding()
	if err != nil {
		return err
	}

	// Generate embeddings for the document
	ctx := context.Background()
	embedding, err := embeddingModel.EmbedQuery(ctx, content)
	if err != nil {
		return err
	}

	// Add the document to the vector store
	err = client.Upsert(ctx, id, embedding, map[string]string{"content": content})
	if err != nil {
		return err
	}

	log.Printf("Document with ID %s added successfully", id)
	return nil
}
