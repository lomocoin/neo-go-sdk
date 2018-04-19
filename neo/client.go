package neo

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/lomocoin/neo-go-sdk/neo/models"
	"github.com/lomocoin/neo-go-sdk/neo/models/response"
)

type (
	// Client is the entrypoint for the package, it is used to carry out all actions.
	Client struct {
		Node     string
		nodeURIs []string
	}
)

// NewClient creates a new Client struct, with a single node URI.
func NewClient(nodeURI string) Client {
	return Client{
		Node:     nodeURI,
		nodeURIs: []string{nodeURI},
	}
}

// NewClientUsingMultipleNodes creates a new Client struct, and allows multiple node URIs
// to be passed in. Before the Client struct is returned, each node is queried to determine
// its block height. The node with the highest block count is chosen.
func NewClientUsingMultipleNodes(nodeURIs []string) (*Client, error) {
	if len(nodeURIs) == 0 {
		return nil, errors.New("Length of 'nodeURIs' argument must be greater than 0")
	}

	client := Client{
		nodeURIs: nodeURIs,
	}

	client.SelectBestNode()
	return &client, nil
}

// GetBestBlockHash returns the hash of the best block in the chain.
func (c Client) GetBestBlockHash() (string, error) {
	var resp response.String

	err := executeRequest("getbestblockhash", nil, c.Node, &resp)
	if err != nil {
		return "", err
	}

	return resp.Result, nil
}

// GetBlockByHash returns the corresponding block information according to the specified
// hash value.
func (c Client) GetBlockByHash(hash string) (*models.Block, error) {
	requestBodyParams := []interface{}{
		hash, 1,
	}
	var resp response.Block

	err := executeRequest("getblock", requestBodyParams, c.Node, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// GetBlockByIndex returns the corresponding block information according to the specified
// index value.
func (c Client) GetBlockByIndex(index int64) (*models.Block, error) {
	requestBodyParams := []interface{}{
		index, 1,
	}
	var resp response.Block

	err := executeRequest("getblock", requestBodyParams, c.Node, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// GetBlockCount returns the number of blocks in the chain.
func (c Client) GetBlockCount() (int64, error) {
	var resp response.Integer

	err := executeRequest("getblockcount", nil, c.Node, &resp)
	if err != nil {
		return 0, err
	}

	return resp.Result, nil
}

// GetBlockHash returns the hash value of the corresponding block based on the specified
// index.
func (c Client) GetBlockHash(index int64) (string, error) {
	requestBodyParams := []interface{}{
		index,
	}
	var resp response.String

	err := executeRequest("getblockhash", requestBodyParams, c.Node, &resp)
	if err != nil {
		return "", err
	}

	return resp.Result, nil
}

// GetConnectionCount returns the current number of connections for the node.
func (c Client) GetConnectionCount() (int64, error) {
	var resp response.Integer

	err := executeRequest("getconnectioncount", nil, c.Node, &resp)
	if err != nil {
		return 0, err
	}

	return resp.Result, nil
}

// GetStorage takes a smart contract hash and a storage key, and returns the storage value
// if available.
func (c Client) GetStorage(scriptHash string, storageKey string) (string, error) {
	requestBodyParams := []interface{}{
		scriptHash, hex.EncodeToString([]byte(storageKey)),
	}
	var resp response.String

	err := executeRequest("getstorage", requestBodyParams, c.Node, &resp)
	if err != nil {
		return "", err
	}

	return resp.Result, nil
}

// GetTransaction returns the corresponding transaction information based on the
// specified hash value.
func (c Client) GetTransaction(hash string) (*models.Transaction, error) {
	requestBodyParams := []interface{}{
		hash, 1,
	}
	var resp response.Transaction

	err := executeRequest("getrawtransaction", requestBodyParams, c.Node, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// GetTransactionOutput returns the corresponding transaction output (change) information
// based on the specified hash and index.
func (c Client) GetTransactionOutput(hash string, index int64) (*models.Vout, error) {
	requestBodyParams := []interface{}{
		hash, index,
	}
	var resp response.Vout

	err := executeRequest("gettxout", requestBodyParams, c.Node, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// GetUnconfirmedTransactions returns a slice of transaction hashes that are all
// unconfirmed transactions that the node has in memory.
func (c Client) GetUnconfirmedTransactions() ([]string, error) {
	var response response.StringArray

	err := executeRequest("getrawmempool", nil, c.Node, &response)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

// SelectBestNode selects the best node to use for RPC calls. If there is a single
// node URI then that will be used. If there are 2 or more then each node is called
// and the block count is compared. The node with the heighest block count is used.
func (c *Client) SelectBestNode() error {
	if len(c.nodeURIs) == 1 {
		c.Node = c.nodeURIs[0]
		return nil
	}

	var bestNode string
	highestBlock := int64(0)

	for _, nodeURI := range c.nodeURIs {
		tempClient := NewClient(nodeURI)

		blockCount, err := tempClient.GetBlockCount()
		if err != nil {
			continue
		}

		if blockCount > highestBlock {
			highestBlock = blockCount
			bestNode = nodeURI
		}
	}

	if bestNode == "" {
		return fmt.Errorf("Unable to communicate with any nodes")
	}

	c.Node = bestNode
	return nil
}

// Ping checks if the node is online.
func (c Client) Ping() bool {
	parsedURI, err := url.Parse(c.Node)
	if err != nil {
		return false
	}

	conn, err := net.Dial("tcp", parsedURI.Host)
	if err != nil {
		return false
	}

	_ = conn.Close()

	return true
}

// ValidateAddress takes a public NEO address and checks if it is valid.
func (c Client) ValidateAddress(address string) (bool, error) {
	requestBodyParams := []interface{}{
		address,
	}
	var resp response.StringMap

	err := executeRequest("validateaddress", requestBodyParams, c.Node, &resp)
	if err != nil {
		return false, err
	}

	if _, ok := resp.Result["address"]; !ok {
		return false, nil
	}

	if _, ok := resp.Result["address"].(string); !ok {
		return false, nil
	}

	if _, ok := resp.Result["isvalid"]; !ok {
		return false, nil
	}

	if _, ok := resp.Result["isvalid"].(bool); !ok {
		return false, nil
	}

	returnedAddress := resp.Result["address"].(string)
	valid := resp.Result["isvalid"].(bool)

	if address == returnedAddress && valid {
		return true, nil
	}

	return false, nil
}

// GetBalance 根据指定的资产编号，返回钱包中对应资产的余额信息
// 执行此命令前需要在 Neo-CLI 节点中打开钱包
func (c Client) GetBalance(assetID string) (balance, confirmed string, err error) {
	requestBodyParams := []interface{}{
		assetID,
	}
	type jd struct {
		Balance   string `json:"balance"`
		Confirmed string `json:"confirmed"`
	}
	var resp struct {
		response.StringMap
		Result jd `json:"result"`
	}

	err = executeRequest("getbalance", requestBodyParams, c.Node, &resp)
	if err != nil {
		return
	}
	balance, confirmed = resp.Result.Balance, resp.Result.Confirmed
	return
}

// GetNewAddress 创建一个新的地址
// 执行此命令前需要在 Neo-CLI 节点中打开钱包
func (c Client) GetNewAddress() (address string, err error) {
	var resp struct {
		response.StringMap
		Result string `json:"result"`
	}

	err = executeRequest("getnewaddress", nil, c.Node, &resp)
	if err != nil {
		return
	}
	address = resp.Result
	return
}

// SendToAddress 向指定地址转账
// 执行此命令前需要在 Neo-CLI 节点中打开钱包
func (c Client) SendToAddress(assetID, toAddress string, amount interface{}) (txID string, err error) {
	requestBodyParams := []interface{}{
		assetID,
		toAddress,
		amount,
	}

	var resp response.Transaction

	err = executeRequest("sendtoaddress", requestBodyParams, c.Node, &resp)
	if err != nil {
		return
	}
	txID = resp.Result.ID
	return
}
