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
	var root []byte
	if node.Type() == FILE {
		root, _ := StoreFile(store, node, h)
		return root
	} else if node.Type() == DIR {
		root := StoreDir(store, node, h)
		return root
	}
	return root
}

// 存文件
func StoreFile(store KVStore, node Node, h hash.Hash) ([]byte, []byte) { //返回hash和data
	file := node.(File).Bytes()
	t := []byte("blob") //小于256*1024为blob
	if node.Size() > 256*1024 {
		t = []byte("list")                //大于256*1024为list
		num := node.Size() / (256 * 1024) //计算有几个256*1024的块
		object := Object{}
		if node.Size()%(256*1024) != 0 {
			num++
		}
		//将块存入KVStore中
		for i := 0; i < int(num); i++ {
			start := i * 256 * 1024
			end := (i + 1) * 256 * 1024
			if end > int(node.Size()) {
				end = int(node.Size())
			}
			sli := file[start:end]
			obj := Object{Data: sli}
			jsonObj, _ := json.Marshal(obj)
			h.Reset()
			key := h.Sum(jsonObj)
			boo, _ := store.Has(key)
			if !boo {
				store.Put(key, jsonObj)
			}
			object.Data = append(object.Data, []byte("blob")...)
			object.Links = append(object.Links, Link{Hash: key, Size: end - start})
		}
		//写入到KVStore
		obj := Object{Data: object.Data, Links: object.Links}
		jsonObj, _ := json.Marshal(obj)
		h.Reset()
		key := h.Sum(jsonObj)
		boo, _ := store.Has(key)
		if !boo {
			store.Put(key, jsonObj)
		}
		return key, t
	} else {
		obj := Object{Data: file}
		jsonObj, _ := json.Marshal(obj)
		h.Reset()
		key := h.Sum(jsonObj)
		boo, _ := store.Has(key)
		if !boo {
			store.Put(key, jsonObj)
		}
		return key, t
	}
}

// 存文件夹
func StoreDir(store KVStore, node Node, h hash.Hash) []byte { //返回hash
	tree := Object{}
	dir := node.(Dir)
	it := dir.It()
	for it.Next() {
		childNode := it.Node()
		if childNode.Type() == FILE {
			key, t := StoreFile(store, childNode, h)
			tree.Data = append(tree.Data, t...)
			tree.Links = append(tree.Links, Link{Name: childNode.Name(), Hash: key, Size: int(childNode.Size())})
		} else if childNode.Type() == DIR {
			key := StoreDir(store, childNode, h)
			tree.Data = append(tree.Data, []byte("tree")...)
			tree.Links = append(tree.Links, Link{Name: childNode.Name(), Hash: key, Size: int(childNode.Size())})
		}
	}
	jsonTree, _ := json.Marshal(tree)
	h.Reset()
	key := h.Sum(jsonTree)
	boo, _ := store.Has(key)
	if !boo {
		store.Put(key, jsonTree)
	}
	return key
}
