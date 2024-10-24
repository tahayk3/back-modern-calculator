package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func handleOperation(w http.ResponseWriter, r *http.Request) {
	// Configurar CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Responder a la solicitud OPTIONS (preflight)
	if r.Method == http.MethodOptions {
		return
	}

	// Procesar la solicitud POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Leer la clave de la API desde la variable de entorno
	apiKey := os.Getenv("API_KEY")
	log.Println("Valor de API_KEY:", apiKey)
	if apiKey == "" {
		log.Fatal("La clave de la API no está configurada")
	}

	// Crear cliente de la API
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Leer la imagen de la solicitud
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error al leer la imagen", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imgData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error al procesar la imagen", http.StatusInternalServerError)
		return
	}

	// Crear el modelo generativo
	model := client.GenerativeModel("gemini-1.5-flash")

	// Preparar el prompt para la generación
	prompt := []genai.Part{
		genai.ImageData("jpeg", imgData),
		genai.Text("¿puedes realizar la operacion que aparece en la imagen?"),
	}

	// Generar el contenido
	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		http.Error(w, "Error al generar contenido", http.StatusInternalServerError)
		return
	}

	// Convertir la respuesta a JSON y enviarla
	respJSON, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Error al serializar la respuesta", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respJSON)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Usa el puerto por defecto si no está configurado
	}

	http.HandleFunc("/operation", handleOperation)

	fmt.Printf("Servidor escuchando en http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
