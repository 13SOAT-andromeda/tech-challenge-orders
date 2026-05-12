package e2e

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/http/handlers"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/http/response"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrder(t *testing.T) {
	cfg, err := SetupTest()

	if err != nil {
		t.Skipf("Error on setup config variables: %v", err)
	}

	apiUrl := GetApiUrl(*cfg)

	healthResp, err := http.Get(apiUrl + "/health")
	if err != nil {
		t.Skipf("Application not running at %s, skipping E2E tests: %v", apiUrl, err)
		return
	}
	healthResp.Body.Close()

	// Identity headers injected by the Lambda Authorizer for the admin user.
	adminEmail := cfg.AdminUser.Email
	adminRole := "administrator"
	adminID := "1"

	t.Run("should create order successfully", func(t *testing.T) {
		var customerID, vehicleID, companyID, customerVehicleID uint
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)

		var createCustomerReq handlers.CreateCustomerRequest

		createCustomerReq.Name = "Cliente Teste Order"
		createCustomerReq.Email = "clienteorder" + timestamp + "@test.com"
		createCustomerReq.Document = GenerateValidCPF(time.Now().UnixNano())
		createCustomerReq.Type = "pf"
		createCustomerReq.Contact = "11999999999"
		createCustomerReq.Address = "Rua Teste"
		createCustomerReq.AddressNumber = "123"
		createCustomerReq.City = "São Paulo"
		createCustomerReq.Neighborhood = "Centro"
		createCustomerReq.Country = "Brasil"
		createCustomerReq.ZipCode = "01000000"

		payload, err := BuildBody(createCustomerReq)
		require.NoError(t, err, "failed to build customer request")

		resp, err := NewIdentifiedReq("POST", apiUrl+"/customers", payload, adminID, adminEmail, adminRole)
		require.NoError(t, err, "failed to create customer")
		defer resp.Body.Close()

		var customerResponse response.BaseResponse[domain.Customer]
		err = ParseBody(resp, &customerResponse)
		require.NoError(t, err, "failed to parse customer response")
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.True(t, customerResponse.Success)
		assert.NotZero(t, customerResponse.Data.ID)
		customerID = customerResponse.Data.ID

		// Create Vehicle
		var createVehicleReq handlers.CreateVehicleRequest

		createVehicleReq.Name = "Corolla XEI 1.8"
		createVehicleReq.Color = "Branco"
		createVehicleReq.Brand = "Toyota"
		createVehicleReq.Plate = GenerateValidPlate(time.Now().UnixNano())
		createVehicleReq.Year = 2020

		payload, err = BuildBody(createVehicleReq)
		require.NoError(t, err, "failed to build vehicle request")

		resp, err = NewIdentifiedReq("POST", apiUrl+"/vehicles", payload, adminID, adminEmail, adminRole)
		require.NoError(t, err, "failed to create vehicle")
		defer resp.Body.Close()

		var vehicleResponse response.BaseResponse[domain.Vehicle]
		err = ParseBody(resp, &vehicleResponse)
		require.NoError(t, err, "failed to parse vehicle response")
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.True(t, vehicleResponse.Success)
		vehicleID = vehicleResponse.Data.ID

		// Associate Vehicle to Customer (creates CustomerVehicle)
		resp, err = NewIdentifiedReq("POST", apiUrl+"/customers/"+strconv.Itoa(int(customerID))+"/vehicles/"+strconv.Itoa(int(vehicleID)), nil, adminID, adminEmail, adminRole)
		require.NoError(t, err, "failed to associate vehicle to customer")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Get CustomerVehicles to get CustomerVehicle ID
		resp, err = NewIdentifiedReq("GET", apiUrl+"/customers/"+strconv.Itoa(int(customerID))+"/vehicles", nil, adminID, adminEmail, adminRole)
		require.NoError(t, err, "failed to get customer vehicles")
		defer resp.Body.Close()

		var customerVehiclesResponse response.BaseResponse[[]domain.CustomerVehicle]
		err = ParseBody(resp, &customerVehiclesResponse)
		require.NoError(t, err, "failed to parse customer vehicles response")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, customerVehiclesResponse.Success)
		require.NotEmpty(t, customerVehiclesResponse.Data, "customer should have vehicles")

		// Find the CustomerVehicle that matches our vehicle
		var foundCustomerVehicle *domain.CustomerVehicle
		for i := range customerVehiclesResponse.Data {
			if customerVehiclesResponse.Data[i].VehicleId == vehicleID {
				foundCustomerVehicle = &customerVehiclesResponse.Data[i]
				break
			}
		}
		require.NotNil(t, foundCustomerVehicle, "vehicle should be associated with customer")
		assert.Equal(t, customerID, foundCustomerVehicle.CustomerId)
		assert.Equal(t, vehicleID, foundCustomerVehicle.VehicleId)
		assert.Equal(t, vehicleID, foundCustomerVehicle.Vehicle.ID)
		customerVehicleID = foundCustomerVehicle.ID

		// Create Company
		var createCompanyReq handlers.CreateCompanyRequest

		createCompanyReq.Name = "Empresa Teste Order"
		createCompanyReq.Email = "empresaorder" + timestamp + "@test.com"
		createCompanyReq.Document = "123" + timestamp[len(timestamp)-4:]
		createCompanyReq.Contact = "11999999999"
		createCompanyReq.Address = "Rua Empresa"
		createCompanyReq.AddressNumber = "456"
		createCompanyReq.City = "São Paulo"
		createCompanyReq.Neighborhood = "Centro"
		createCompanyReq.Country = "Brasil"
		createCompanyReq.ZipCode = "01000000"

		payload, err = BuildBody(createCompanyReq)
		require.NoError(t, err, "failed to build company request")

		resp, err = NewIdentifiedReq("POST", apiUrl+"/companies", payload, adminID, adminEmail, adminRole)
		require.NoError(t, err, "failed to create company")
		defer resp.Body.Close()

		var companyResponse response.BaseResponse[domain.Company]
		err = ParseBody(resp, &companyResponse)
		require.NoError(t, err, "failed to parse company response")
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.True(t, companyResponse.Success)
		assert.NotZero(t, companyResponse.Data.ID)
		companyID = companyResponse.Data.ID

		// Create Product
		var createProductReq handlers.CreateProductRequest

		createProductReq.Name = "Produto Teste Order " + timestamp
		createProductReq.Price = 10000
		createProductReq.Stock = 100

		payload, err = BuildBody(createProductReq)
		require.NoError(t, err, "failed to build product request")

		resp, err = NewIdentifiedReq("POST", apiUrl+"/products", payload, adminID, adminEmail, adminRole)
		require.NoError(t, err, "failed to create product")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Create Maintenance
		var createMaintenanceReq handlers.CreateMaintenanceRequest

		createMaintenanceReq.Name = "Manutenção Teste Order " + timestamp
		createMaintenanceReq.Price = 500.00
		createMaintenanceReq.CategoryId = "standard"

		payload, err = BuildBody(createMaintenanceReq)
		require.NoError(t, err, "failed to build maintenance request")

		resp, err = NewIdentifiedReq("POST", apiUrl+"/maintenances", payload, adminID, adminEmail, adminRole)
		require.NoError(t, err, "failed to create maintenance")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Now create the Order
		var createOrderReq handlers.CreateOrderRequest

		note := "Nota de teste para a ordem"
		createOrderReq.VehicleKilometers = 50000
		createOrderReq.Note = &note
		createOrderReq.CustomerVehicleID = customerVehicleID
		createOrderReq.CompanyID = companyID

		payload, err = BuildBody(createOrderReq)
		require.NoError(t, err, "failed to build order request")

		resp, err = NewIdentifiedReq("POST", apiUrl+"/orders", payload, adminID, adminEmail, adminRole)
		require.NoError(t, err, "failed to create order")
		defer resp.Body.Close()

		var orderResponse response.BaseResponse[domain.Order]
		err = ParseBody(resp, &orderResponse)
		require.NoError(t, err, "failed to parse order response")

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.True(t, orderResponse.Success)
		assert.NotZero(t, orderResponse.Data.ID)
		assert.Equal(t, createOrderReq.VehicleKilometers, orderResponse.Data.VehicleKilometers)
		assert.Equal(t, *createOrderReq.Note, *orderResponse.Data.Note)
		assert.Equal(t, domain.OrderStatuses.RECEIVED, orderResponse.Data.Status)
		assert.Equal(t, customerVehicleID, orderResponse.Data.CustomerVehicleID)
		assert.Equal(t, companyID, orderResponse.Data.CompanyID)
	})
}
