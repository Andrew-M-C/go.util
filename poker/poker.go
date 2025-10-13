// Package poker 简单封装一个扑克牌库, 自己写着玩
package poker

import "fmt"

// Card 表示一张扑克牌
type Card struct {
	suit  Suit
	point Point
}

// NewCard 初始化一个普通牌
func NewCard(suit Suit, point Point) Card {
	return Card{
		suit:  suit,
		point: point,
	}
}

// NewRedJoker 初始化一个大王
func NewRedJoker() Card {
	return Card{
		point: RedJoker,
	}
}

// NewBlackJoker 初始化一个小王
func NewBlackJoker() Card {
	return Card{
		point: BlackJoker,
	}
}

// GetSuit 获取扑克牌的花色。Joker 牌没有花色
func (c Card) GetSuit() Suit {
	return c.suit
}

// GetPoint 获取扑克牌的点数。Joker 牌在点数上对应 BlackJoker 和 RedJoker
func (c Card) GetPoint() Point {
	return c.point
}

// IsRedJoker 判断是否是红鬼牌
func (c Card) IsRedJoker() bool {
	return c.point == RedJoker
}

// IsBlackJoker 判断是否是黑鬼牌
func (c Card) IsBlackJoker() bool {
	return c.point == BlackJoker
}

// IsJoker 判断是否是鬼牌
func (c Card) IsJoker() bool {
	return c.IsRedJoker() || c.IsBlackJoker()
}

func (c Card) String() string {
	switch c.point {
	case RedJoker:
		return "🔴🃏"
	case BlackJoker:
		return "⚫🃏"
	case Jack:
		return string(c.suit) + "J"
	case Queen:
		return string(c.suit) + "Q"
	case King:
		return string(c.suit) + "Q"
	case Ace:
		return string(c.suit) + "A"
	default:
		return fmt.Sprintf("%s%d", c.suit, c.point)
	}
}

type Suit string

const (
	Diamond Suit = "♦️"
	Clubs   Suit = "♣️"
	Heart   Suit = "♥️"
	Spade   Suit = "♠️"
)

type Point int8

const (
	Ace        Point = 1
	Jack       Point = 11
	Queen      Point = 12
	King       Point = 13
	BlackJoker Point = 14
	RedJoker   Point = 15
)
