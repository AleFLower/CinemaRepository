package utils

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapGRPCErrorToUser traduce gli errori tecnici gRPC in messaggi comprensibili
func MapGRPCErrorToUser(err error) string {
	if err == nil {
		return ""
	}

	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.Unavailable:
			return "Il servizio è momentaneamente non raggiungibile. Riprova tra poco."
		case codes.DeadlineExceeded:
			return "La richiesta ha impiegato troppo tempo. Controlla la tua connessione."
		case codes.NotFound:
			return "L'elemento richiesto non è stato trovato."
		case codes.Unimplemented:
			return "Funzionalità non ancora disponibile."
		}
		// Ritorna il messaggio d'errore inviato dal server (es. "Posto occupato")
		return st.Message()
	}

	return "Si è verificato un errore imprevisto. I nostri tecnici sono al lavoro."
}
