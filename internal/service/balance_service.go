package service

import (
	"context"
	"fmt"
	"github.com/bersh/alluvial_test_1/internal/metrics"
	"math/big"
	"sort"

	"github.com/bersh/alluvial_test_1/internal/client"
)

// BalanceService handles balance-related operations
type BalanceService struct {
	clientPool client.Pool
}

// NewBalanceService creates a new balance service
func NewBalanceService(clientPool client.Pool) *BalanceService {
	return &BalanceService{
		clientPool: clientPool,
	}
}

// GetBalance retrieves a balance from multiple clients and returns the consensus result
func (s *BalanceService) GetBalance(ctx context.Context, address, blockParam string) (*big.Int, error) {
	responses, err := s.clientPool.QueryBalanceFromAllClients(ctx, address, blockParam)
	if err != nil {
		return nil, fmt.Errorf("failed to query balances: %w", err)
	}

	consensusBalance, hasDiscrepancy := getConsensusBalance(responses, address)

	if hasDiscrepancy {
		fmt.Printf("Balance discrepancy detected for address %s\n", address)
	}

	return consensusBalance, nil
}

// getConsensusBalance determines the most reliable balance from multiple client responses
func getConsensusBalance(responses []client.BalanceResponse, address string) (*big.Int, bool) {
	if len(responses) == 1 {
		return responses[0].Balance, false
	}

	// Track the original order of balances to ensure deterministic behavior
	// when there are equal vote counts
	balanceOrder := make(map[string]int)
	balanceCounts := make(map[string]int)
	balanceMap := make(map[string]*big.Int)

	for i, resp := range responses {
		balanceStr := resp.Balance.String()
		// Only record the first occurrence position for each balance
		if _, exists := balanceOrder[balanceStr]; !exists {
			balanceOrder[balanceStr] = i
		}
		balanceCounts[balanceStr]++
		balanceMap[balanceStr] = resp.Balance
	}

	type balanceCount struct {
		balance string
		count   int
		order   int // Original order in the responses
	}

	sortedBalances := make([]balanceCount, 0, len(balanceCounts))
	for balance, count := range balanceCounts {
		sortedBalances = append(sortedBalances, balanceCount{
			balance: balance,
			count:   count,
			order:   balanceOrder[balance],
		})
	}

	// Sort by count (descending) and then by original order (ascending) for stable results
	sort.Slice(sortedBalances, func(i, j int) bool {
		if sortedBalances[i].count != sortedBalances[j].count {
			return sortedBalances[i].count > sortedBalances[j].count
		}
		// If counts are equal, use the original order to ensure deterministic behavior
		return sortedBalances[i].order < sortedBalances[j].order
	})

	hasDiscrepancy := len(balanceCounts) > 1

	if hasDiscrepancy {
		metrics.RecordBalanceDiscrepancy(address)
	}

	consensusBalance := balanceMap[sortedBalances[0].balance]
	return consensusBalance, hasDiscrepancy
}
