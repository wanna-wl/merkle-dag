package merkledag

import (
	"encoding/json"
	"hash"
)

type Link struct {
	Name string
	Hash []byte
	Size int
}

type Object struct {
	Links []Link
	Data  []byte
}

func Add(store KVStore, node Node, h hash.Hash) []byte {
	// TODO 将分片写入到KVStore中，并返回Merkle Root
	switch n := node.(type) {
	case File:
		data := n.Bytes()
		hashBytes := h.Sum(data)
		//保存到store中
		err := store.Put(hashBytes, data)
		if err != nil {
			panic(err)
		}
		return hashBytes
	case Dir:
		links := make([]Link, 0)
		it := n.It()
		for it.Next() {
			childNode := it.Node()
			childHash := Add(store, childNode, h)
			// 创建并添加链接到当前节点的Links字段中
			links = append(links, Link{
				Name: childNode.Name(),
				Hash: childHash,
				Size: int(childNode.Size()),
			})
		}
		object := Object{
			Links: links,
			Data:  nil,
		}
		objectBytes, err1 := json.Marshal(object)
		if err1 != nil {
			panic(err1)
		}
		objectHash := h.Sum(objectBytes)
		//保存到store中
		err2 := store.Put(objectHash, objectBytes)
		if err2 != nil {
			panic(err2)
		}
		return objectHash
	default:
		return nil
	}
}
