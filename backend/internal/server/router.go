package server

import (
	"clinic.local/backend/internal/handler"
	"clinic.local/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(
	patientService service.PatientCreator,
	catalogService service.CatalogReader,
	scheduleService service.ScheduleReader,
	appointmentService service.AppointmentManager,
	authService service.Authenticator,
) *gin.Engine {
	router := gin.Default()

	authHandler := handler.NewAuthHandler(authService)
	healthHandler := handler.NewHealthHandler()
	patientHandler := handler.NewPatientHandler(patientService)
	catalogHandler := handler.NewCatalogHandler(catalogService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)
	appointmentHandler := handler.NewAppointmentHandler(appointmentService)

	api := router.Group("/api/v1")
	{
		api.GET("/health", healthHandler.GetHealth)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/patient/register", authHandler.RegisterPatient)
		api.POST("/patients", patientHandler.CreatePatient)
		api.GET("/services/pricelist", catalogHandler.GetPriceList)
		api.GET("/doctors/:doctorId/schedule", scheduleHandler.GetDoctorSchedule)
		api.POST("/appointments", appointmentHandler.BookAppointment)
		api.POST("/appointments/:appointmentId/cancel", appointmentHandler.CancelAppointment)
		api.POST("/appointments/:appointmentId/complete", appointmentHandler.CompleteVisit)
		api.POST("/invoices/:invoiceId/payments", appointmentHandler.RegisterPayment)
	}

	return router
}
