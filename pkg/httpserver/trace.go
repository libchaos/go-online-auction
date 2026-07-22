package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Trace starts a server span per request using the global tracer provider set
// up by the telemetry module. The span name is updated to the matched route
// pattern after routing so spans aggregate across concrete IDs (e.g.
// "/auctions/{id}" instead of "/auctions/42"). The context carrying the span is
// passed downstream via r.WithContext; because tracer.Start embeds the parent
// context, chi's route context is preserved.
func Trace(next http.Handler) http.Handler {
	tracer := otel.Tracer("auction/http")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), r.Method+" "+routeLabel(r),
			trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		next.ServeHTTP(w, r.WithContext(ctx))

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = routeLabel(r)
		} else {
			route = r.Method + " " + route
		}
		span.SetName(route)
		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", route),
			attribute.String("http.path", r.URL.Path),
		)
	})
}
