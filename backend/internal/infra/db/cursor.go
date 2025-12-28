package db

import (
	"errors"

	"gorm.io/gorm"
)

func (m *MySQL) GetCursor(
	chainID uint64,
	contractAddr string,
	eventName string,
) (uint64, error) {

	var c SyncCursor
	err := m.DB.Where(
		"chain_id = ? AND contract_address = ? AND event_name = ?",
		chainID, contractAddr, eventName,
	).First(&c).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, ErrCursorNotFound // 不再返回 nil
	}
	if err != nil {
		return 0, err
	}
	return c.LastBlock, nil
}

func (m *MySQL) UpsertCursor(chainID uint64, contractAddr, EventName string, lastBlock uint64) error {
	var c SyncCursor
	err := m.DB.Where("chain_id = ? AND contract_address = ? AND event_name = ?",
		chainID, contractAddr, EventName).First(&c).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return m.DB.Create(&SyncCursor{
			ChainID: chainID, ContractAddress: contractAddr, EventName: EventName, LastBlock: lastBlock,
		}).Error
	}
	if err != nil {
		return err
	}

	c.LastBlock = lastBlock
	return m.DB.Save(&c).Error
}
