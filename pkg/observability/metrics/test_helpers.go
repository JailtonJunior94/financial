package metrics

import "github.com/JailtonJunior94/devkit-go/pkg/observability/fake"

// NewTestCardMetrics cria uma instância de CardMetrics para testes
// usando um fake provider para evitar dependências de exportação
func NewTestCardMetrics() *CardMetrics {
	fakeProvider := fake.NewProvider()
	return NewCardMetrics(fakeProvider)
}
