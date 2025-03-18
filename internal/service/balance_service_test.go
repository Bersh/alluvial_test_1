package service

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/bersh/alluvial_test_1/internal/client"
	"github.com/bersh/alluvial_test_1/internal/client/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBalanceService_GetBalance(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  []client.BalanceResponse
		mockError      error
		expectedResult *big.Int
		expectedError  bool
	}{
		{
			name: "Successful balance query with consensus",
			mockResponses: []client.BalanceResponse{
				{
					ClientName: "client1",
					Balance:    big.NewInt(1000),
					Error:      nil,
				},
				{
					ClientName: "client2",
					Balance:    big.NewInt(1000),
					Error:      nil,
				},
			},
			mockError:      nil,
			expectedResult: big.NewInt(1000),
			expectedError:  false,
		},
		{
			name: "Successful balance query with discrepancy",
			mockResponses: []client.BalanceResponse{
				{
					ClientName: "client1",
					Balance:    big.NewInt(1000),
					Error:      nil,
				},
				{
					ClientName: "client2",
					Balance:    big.NewInt(2000),
					Error:      nil,
				},
				{
					ClientName: "client3",
					Balance:    big.NewInt(1000),
					Error:      nil,
				},
			},
			mockError:      nil,
			expectedResult: big.NewInt(1000),
			expectedError:  false,
		},
		{
			name:           "Error when querying clients",
			mockResponses:  nil,
			mockError:      errors.New("no clients available"),
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool := new(mocks.Pool)

			mockPool.On("QueryBalanceFromAllClients",
				mock.Anything,                 // context
				mock.AnythingOfType("string"), // address
				mock.AnythingOfType("string"), // blockParam
			).Return(tt.mockResponses, tt.mockError)

			service := NewBalanceService(mockPool)

			result, err := service.GetBalance(context.Background(), "0x123", "latest")

			mockPool.AssertExpectations(t)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 0, tt.expectedResult.Cmp(result),
					"Expected balance %s, got %s", tt.expectedResult.String(), result.String())
			}
		})
	}
}

func TestGetConsensusBalance(t *testing.T) {
	tests := []struct {
		name           string
		responses      []client.BalanceResponse
		expectedValue  *big.Int
		hasDiscrepancy bool
	}{
		{
			name: "Single response returns that value",
			responses: []client.BalanceResponse{
				{
					ClientName: "client1",
					Balance:    big.NewInt(100),
					Error:      nil,
				},
			},
			expectedValue:  big.NewInt(100),
			hasDiscrepancy: false,
		},
		{
			name: "Multiple identical responses returns consensus",
			responses: []client.BalanceResponse{
				{
					ClientName: "client1",
					Balance:    big.NewInt(100),
					Error:      nil,
				},
				{
					ClientName: "client2",
					Balance:    big.NewInt(100),
					Error:      nil,
				},
				{
					ClientName: "client3",
					Balance:    big.NewInt(100),
					Error:      nil,
				},
			},
			expectedValue:  big.NewInt(100),
			hasDiscrepancy: false,
		},
		{
			name: "Majority vote with discrepancy",
			responses: []client.BalanceResponse{
				{
					ClientName: "client1",
					Balance:    big.NewInt(100),
					Error:      nil,
				},
				{
					ClientName: "client2",
					Balance:    big.NewInt(100),
					Error:      nil,
				},
				{
					ClientName: "client3",
					Balance:    big.NewInt(200),
					Error:      nil,
				},
			},
			expectedValue:  big.NewInt(100),
			hasDiscrepancy: true,
		},
		{
			name: "Equal votes chooses first",
			responses: []client.BalanceResponse{
				{
					ClientName: "client1",
					Balance:    big.NewInt(100),
					Error:      nil,
				},
				{
					ClientName: "client2",
					Balance:    big.NewInt(200),
					Error:      nil,
				},
			},
			expectedValue:  big.NewInt(100),
			hasDiscrepancy: true,
		},
		{
			name: "Complex scenario with multiple different responses",
			responses: []client.BalanceResponse{
				{
					ClientName: "client1",
					Balance:    big.NewInt(100),
					Error:      nil,
				},
				{
					ClientName: "client2",
					Balance:    big.NewInt(100),
					Error:      nil,
				},
				{
					ClientName: "client3",
					Balance:    big.NewInt(200),
					Error:      nil,
				},
				{
					ClientName: "client4",
					Balance:    big.NewInt(300),
					Error:      nil,
				},
				{
					ClientName: "client5",
					Balance:    big.NewInt(100),
					Error:      nil,
				},
			},
			expectedValue:  big.NewInt(100),
			hasDiscrepancy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, hasDiscrepancy := getConsensusBalance(tt.responses, "0x123")

			assert.Equal(t, 0, result.Cmp(tt.expectedValue),
				"Expected balance %s, got %s", tt.expectedValue.String(), result.String())
			assert.Equal(t, tt.hasDiscrepancy, hasDiscrepancy)
		})
	}
}
