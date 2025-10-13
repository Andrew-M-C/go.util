package poker

import (
	"math/rand/v2"
)

// Deck 表示一副牌
type Deck struct {
	cards []Card
}

// NewDeck 初始化一个全新的扑克牌堆, 排序顺序为: 大王、小王、黑桃KQJ...32A、红桃、梅花、方块
func NewDeck() *Deck {
	res := make([]Card, 0, 13*4+2)
	res = append(res, NewRedJoker())
	res = append(res, NewBlackJoker())

	suits := []Suit{Spade, Heart, Clubs, Diamond}
	for _, suit := range suits {
		for i := Ace; i <= King; i++ {
			res = append(res, NewCard(suit, i))
		}
	}

	return &Deck{
		cards: res,
	}
}

// Shuffle 随机洗牌
func (d *Deck) Shuffle() {
	rand.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
}

// Len 返回长度
func (d *Deck) Len() int {
	return len(d.cards)
}

// Draw 抽牌
func (d *Deck) Draw(n int) []Card {
	if len(d.cards) < n {
		return nil
	}
	res := d.cards[:n]
	d.cards = d.cards[n:]
	return res
}

// FaroShuffle 法罗洗牌, 也就是完美洗牌 (Out-shuffle 版本)
// 将牌堆分成两半, 然后交错叠放。第一张和最后一张保持原位。
// 对于 52 张牌的标准牌堆, 执行 8 次 FaroShuffle 可以恢复原始顺序
func (d *Deck) FaroShuffle() {
	n := len(d.cards)
	if n <= 1 {
		return
	}

	// 创建新的切片存储洗牌后的结果
	result := make([]Card, n)

	// 计算中点 - 对于奇数张牌, 后半部分会多一张
	mid := n / 2

	// Out-shuffle: 前后交错放置
	// 前半部分: cards[0:mid]
	// 后半部分: cards[mid:n]
	for i := 0; i < mid; i++ {
		result[2*i] = d.cards[i]
		result[2*i+1] = d.cards[mid+i]
	}

	// 如果是奇数张牌, 最后一张保持在末尾
	if n%2 == 1 {
		result[n-1] = d.cards[n-1]
	}

	d.cards = result
}

// FaroShuffleIn 法罗洗牌的 In-shuffle 版本
// 第一张和最后一张牌会移动到内部
func (d *Deck) FaroShuffleIn() {
	n := len(d.cards)
	if n <= 1 {
		return
	}

	// 创建新的切片存储洗牌后的结果
	result := make([]Card, n)

	// 计算中点
	mid := n / 2

	// In-shuffle: 从后半部分第一张开始交错放置
	// 前半部分: cards[0:mid]
	// 后半部分: cards[mid:n]
	for i := 0; i < mid; i++ {
		result[2*i] = d.cards[mid+i]
		result[2*i+1] = d.cards[i]
	}

	// 如果是奇数张牌, 最后一张放在末尾
	if n%2 == 1 {
		result[n-1] = d.cards[n-1]
	}

	d.cards = result
}
