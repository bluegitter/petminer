package utils

import (
	"encoding/json"
	"miningpet/internal/models"
	"sync"
)

// ObjectPool 对象池管理器
type ObjectPool struct {
	// 事件对象池
	eventPool sync.Pool
	// 宠物对象池
	petPool sync.Pool
	// JSON编码器池
	encoderPool sync.Pool
	// JSON解码器池
	decoderPool sync.Pool
	// 字节缓冲池
	bufferPool sync.Pool
}

// NewObjectPool 创建对象池
func NewObjectPool() *ObjectPool {
	pool := &ObjectPool{}
	
	// 初始化事件对象池
	pool.eventPool.New = func() interface{} {
		return &models.Event{}
	}
	
	// 初始化宠物对象池
	pool.petPool.New = func() interface{} {
		return &models.Pet{}
	}
	
	// 初始化JSON编码器池
	pool.encoderPool.New = func() interface{} {
		return json.NewEncoder(nil)
	}
	
	// 初始化JSON解码器池
	pool.decoderPool.New = func() interface{} {
		return json.NewDecoder(nil)
	}
	
	// 初始化字节缓冲池
	pool.bufferPool.New = func() interface{} {
		return make([]byte, 0, 1024) // 预分配1KB
	}
	
	return pool
}

// GetEvent 获取事件对象
func (p *ObjectPool) GetEvent() *models.Event {
	event := p.eventPool.Get().(*models.Event)
	// 重置事件对象
	*event = models.Event{}
	return event
}

// PutEvent 归还事件对象
func (p *ObjectPool) PutEvent(event *models.Event) {
	if event != nil {
		p.eventPool.Put(event)
	}
}

// GetPet 获取宠物对象
func (p *ObjectPool) GetPet() *models.Pet {
	pet := p.petPool.Get().(*models.Pet)
	// 重置宠物对象
	*pet = models.Pet{}
	return pet
}

// PutPet 归还宠物对象
func (p *ObjectPool) PutPet(pet *models.Pet) {
	if pet != nil {
		p.petPool.Put(pet)
	}
}

// GetEncoder 获取JSON编码器
func (p *ObjectPool) GetEncoder() *json.Encoder {
	return p.encoderPool.Get().(*json.Encoder)
}

// PutEncoder 归还JSON编码器
func (p *ObjectPool) PutEncoder(encoder *json.Encoder) {
	if encoder != nil {
		p.encoderPool.Put(encoder)
	}
}

// GetDecoder 获取JSON解码器
func (p *ObjectPool) GetDecoder() *json.Decoder {
	return p.decoderPool.Get().(*json.Decoder)
}

// PutDecoder 归还JSON解码器
func (p *ObjectPool) PutDecoder(decoder *json.Decoder) {
	if decoder != nil {
		p.decoderPool.Put(decoder)
	}
}

// GetBuffer 获取字节缓冲
func (p *ObjectPool) GetBuffer() []byte {
	return p.bufferPool.Get().([]byte)
}

// PutBuffer 归还字节缓冲
func (p *ObjectPool) PutBuffer(buffer []byte) {
	if buffer != nil {
		// 重置长度但保留容量
		buffer = buffer[:0]
		// 如果缓冲区太大，不要放回池中
		if cap(buffer) <= 64*1024 { // 最大64KB
			p.bufferPool.Put(buffer)
		}
	}
}

// 全局对象池实例
var GlobalPool = NewObjectPool()

// EventWithPool 使用对象池的事件创建函数
func EventWithPool() *models.Event {
	return GlobalPool.GetEvent()
}

// ReleaseEvent 释放事件到对象池
func ReleaseEvent(event *models.Event) {
	GlobalPool.PutEvent(event)
}

// PetWithPool 使用对象池的宠物创建函数
func PetWithPool() *models.Pet {
	return GlobalPool.GetPet()
}

// ReleasePet 释放宠物到对象池
func ReleasePet(pet *models.Pet) {
	GlobalPool.PutPet(pet)
}