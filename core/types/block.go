package types

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

type Header struct {
	// Hash to the previous block
	ParentHash common.Hash
	// Uncles of this block
	UncleHash common.Hash
	// The coin base address
	Coinbase common.Address
	// Block Trie state
	Root common.Hash
	// Tx sha
	TxHash common.Hash
	// Receipt sha
	ReceiptHash common.Hash
	// Bloom
	Bloom Bloom
	// Difficulty for the current block
	Difficulty *big.Int
	// The block number
	Number *big.Int
	// Gas limit
	GasLimit *big.Int
	// Gas used
	GasUsed *big.Int
	// Creation time
	Time uint64
	// Extra data
	Extra string
	// Mix digest for quick checking to prevent DOS
	MixDigest common.Hash
	// Nonce
	Nonce [8]byte
}

func (self *Header) rlpData(withNonce bool) []interface{} {
	fields := []interface{}{
		self.ParentHash,
		self.UncleHash,
		self.Coinbase,
		self.Root,
		self.TxHash,
		self.ReceiptHash,
		self.Bloom,
		self.Difficulty,
		self.Number,
		self.GasLimit,
		self.GasUsed,
		self.Time,
		self.Extra,
	}
	if withNonce {
		fields = append(fields, self.MixDigest, self.Nonce)
	}

	return fields
}

func (self *Header) RlpData() interface{} {
	return self.rlpData(true)
}

func (self *Header) Hash() common.Hash {
	return common.BytesToHash(crypto.Sha3(common.Encode(self.rlpData(true))))
}

func (self *Header) HashNoNonce() common.Hash {
	return common.BytesToHash(crypto.Sha3(common.Encode(self.rlpData(false))))
}

type Block struct {
	// Preset Hash for mock (Tests)
	HeaderHash       common.Hash
	ParentHeaderHash common.Hash
	// ^^^^ ignore ^^^^

	header       *Header
	uncles       []*Header
	transactions Transactions
	Td           *big.Int

	receipts Receipts
	Reward   *big.Int
}

func NewBlock(parentHash common.Hash, coinbase common.Address, root common.Hash, difficulty *big.Int, nonce uint64, extra string) *Block {
	header := &Header{
		Root:       root,
		ParentHash: parentHash,
		Coinbase:   coinbase,
		Difficulty: difficulty,
		Time:       uint64(time.Now().Unix()),
		Extra:      extra,
		GasUsed:    new(big.Int),
		GasLimit:   new(big.Int),
	}
	header.SetNonce(nonce)

	block := &Block{header: header, Reward: new(big.Int)}

	return block
}

func (self *Header) SetNonce(nonce uint64) {
	binary.BigEndian.PutUint64(self.Nonce[:], nonce)
}

func NewBlockWithHeader(header *Header) *Block {
	return &Block{header: header}
}

func (self *Block) DecodeRLP(s *rlp.Stream) error {
	var extblock struct {
		Header *Header
		Txs    []*Transaction
		Uncles []*Header
		TD     *big.Int // optional
	}
	if err := s.Decode(&extblock); err != nil {
		return err
	}
	self.header = extblock.Header
	self.uncles = extblock.Uncles
	self.transactions = extblock.Txs
	self.Td = extblock.TD
	return nil
}

func (self *Block) Header() *Header {
	return self.header
}

func (self *Block) Uncles() []*Header {
	return self.uncles
}

func (self *Block) SetUncles(uncleHeaders []*Header) {
	self.uncles = uncleHeaders
	self.header.UncleHash = common.BytesToHash(crypto.Sha3(common.Encode(uncleHeaders)))
}

func (self *Block) Transactions() Transactions {
	return self.transactions
}

func (self *Block) Transaction(hash common.Hash) *Transaction {
	for _, transaction := range self.transactions {
		if transaction.Hash() == hash {
			return transaction
		}
	}
	return nil
}

func (self *Block) SetTransactions(transactions Transactions) {
	self.transactions = transactions
	self.header.TxHash = DeriveSha(transactions)
}
func (self *Block) AddTransaction(transaction *Transaction) {
	self.transactions = append(self.transactions, transaction)
	self.SetTransactions(self.transactions)
}

func (self *Block) Receipts() Receipts {
	return self.receipts
}

func (self *Block) SetReceipts(receipts Receipts) {
	self.receipts = receipts
	self.header.ReceiptHash = DeriveSha(receipts)
	self.header.Bloom = CreateBloom(receipts)
}
func (self *Block) AddReceipt(receipt *Receipt) {
	self.receipts = append(self.receipts, receipt)
	self.SetReceipts(self.receipts)
}

func (self *Block) RlpData() interface{} {
	return []interface{}{self.header, self.transactions, self.uncles}
}

func (self *Block) RlpDataForStorage() interface{} {
	return []interface{}{self.header, self.transactions, self.uncles, self.Td /* TODO receipts */}
}

// Header accessors (add as you need them)
func (self *Block) Number() *big.Int       { return self.header.Number }
func (self *Block) NumberU64() uint64      { return self.header.Number.Uint64() }
func (self *Block) MixDigest() common.Hash { return self.header.MixDigest }
func (self *Block) Nonce() uint64 {
	return binary.BigEndian.Uint64(self.header.Nonce[:])
}
func (self *Block) SetNonce(nonce uint64) {
	self.header.SetNonce(nonce)
}

func (self *Block) Bloom() Bloom             { return self.header.Bloom }
func (self *Block) Coinbase() common.Address { return self.header.Coinbase }
func (self *Block) Time() int64              { return int64(self.header.Time) }
func (self *Block) GasLimit() *big.Int       { return self.header.GasLimit }
func (self *Block) GasUsed() *big.Int        { return self.header.GasUsed }
func (self *Block) Root() common.Hash        { return self.header.Root }
func (self *Block) SetRoot(root common.Hash) { self.header.Root = root }
func (self *Block) Size() common.StorageSize { return common.StorageSize(len(common.Encode(self))) }
func (self *Block) GetTransaction(i int) *Transaction {
	if len(self.transactions) > i {
		return self.transactions[i]
	}
	return nil
}
func (self *Block) GetUncle(i int) *Header {
	if len(self.uncles) > i {
		return self.uncles[i]
	}
	return nil
}

// Implement pow.Block
func (self *Block) Difficulty() *big.Int     { return self.header.Difficulty }
func (self *Block) HashNoNonce() common.Hash { return self.header.HashNoNonce() }

func (self *Block) Hash() common.Hash {
	if (self.HeaderHash != common.Hash{}) {
		return self.HeaderHash
	} else {
		return self.header.Hash()
	}
}

func (self *Block) ParentHash() common.Hash {
	if (self.ParentHeaderHash != common.Hash{}) {
		return self.ParentHeaderHash
	} else {
		return self.header.ParentHash
	}
}

func (self *Block) String() string {
	return fmt.Sprintf(`BLOCK(%x): Size: %v TD: %v {
NoNonce: %x
Header:
[
%v
]
Transactions:
%v
Uncles:
%v
}
`, self.header.Hash(), self.Size(), self.Td, self.header.HashNoNonce(), self.header, self.transactions, self.uncles)
}

func (self *Header) String() string {
	return fmt.Sprintf(`
	ParentHash:	    %x
	UncleHash:	    %x
	Coinbase:	    %x
	Root:		    %x
	TxSha		    %x
	ReceiptSha:	    %x
	Bloom:		    %x
	Difficulty:	    %v
	Number:		    %v
	GasLimit:	    %v
	GasUsed:	    %v
	Time:		    %v
	Extra:		    %v
	MixDigest:          %x
	Nonce:		    %x`,
		self.ParentHash, self.UncleHash, self.Coinbase, self.Root, self.TxHash, self.ReceiptHash, self.Bloom, self.Difficulty, self.Number, self.GasLimit, self.GasUsed, self.Time, self.Extra, self.MixDigest, self.Nonce)
}

type Blocks []*Block

type BlockBy func(b1, b2 *Block) bool

func (self BlockBy) Sort(blocks Blocks) {
	bs := blockSorter{
		blocks: blocks,
		by:     self,
	}
	sort.Sort(bs)
}

type blockSorter struct {
	blocks Blocks
	by     func(b1, b2 *Block) bool
}

func (self blockSorter) Len() int { return len(self.blocks) }
func (self blockSorter) Swap(i, j int) {
	self.blocks[i], self.blocks[j] = self.blocks[j], self.blocks[i]
}
func (self blockSorter) Less(i, j int) bool { return self.by(self.blocks[i], self.blocks[j]) }

func Number(b1, b2 *Block) bool { return b1.Header().Number.Cmp(b2.Header().Number) < 0 }
