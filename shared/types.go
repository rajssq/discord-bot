package shared

import "net/http"

// Interface HTTPClient define um contrato para realizar requisições HTTP.
// Essa interface permite a substituição do cliente HTTP padrão por uma implementação personalizada,
// facilitando a criação de testes mock para componentes que realizam requisições.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error) // Executa a requisição HTTP e retorna a resposta ou um erro.
}
