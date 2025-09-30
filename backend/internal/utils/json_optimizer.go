package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"miningpet/internal/models"
	"reflect"
	"sync"
)

// JSONOptimizer JSON序列化优化器
type JSONOptimizer struct {
	bufferPool sync.Pool
	
	// 预编译的JSON标签缓存
	tagCache sync.Map
	
	// 常用字段的预编译模板
	eventTemplate    []byte
	petTemplate      []byte
	responseTemplate []byte
}

// NewJSONOptimizer 创建JSON优化器
func NewJSONOptimizer() *JSONOptimizer {
	optimizer := &JSONOptimizer{}
	
	// 初始化缓冲池
	optimizer.bufferPool.New = func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 512))
	}
	
	// 预编译常用模板
	optimizer.precompileTemplates()
	
	return optimizer
}

// precompileTemplates 预编译JSON模板
func (jo *JSONOptimizer) precompileTemplates() {
	// 事件模板
	eventSample := models.Event{
		ID:        "template",
		PetID:     "template",
		PetName:   "template",
		Type:      models.EventExplore,
		Message:   "template",
		Timestamp: models.Event{}.Timestamp,
	}
	jo.eventTemplate, _ = json.Marshal(eventSample)
	
	// 宠物模板
	petSample := models.Pet{
		ID:    "template",
		Name:  "template",
		Owner: "template",
	}
	jo.petTemplate, _ = json.Marshal(petSample)
}

// getBuffer 获取缓冲区
func (jo *JSONOptimizer) getBuffer() *bytes.Buffer {
	return jo.bufferPool.Get().(*bytes.Buffer)
}

// putBuffer 归还缓冲区
func (jo *JSONOptimizer) putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	jo.bufferPool.Put(buf)
}

// FastMarshalEvent 快速序列化事件
func (jo *JSONOptimizer) FastMarshalEvent(event *models.Event) ([]byte, error) {
	buf := jo.getBuffer()
	defer jo.putBuffer(buf)
	
	// 使用预分配的缓冲区进行序列化
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(event); err != nil {
		return nil, err
	}
	
	// 复制结果（因为buf会被重用）
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	
	return result, nil
}

// FastMarshalPet 快速序列化宠物
func (jo *JSONOptimizer) FastMarshalPet(pet *models.Pet) ([]byte, error) {
	buf := jo.getBuffer()
	defer jo.putBuffer(buf)
	
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(pet); err != nil {
		return nil, err
	}
	
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	
	return result, nil
}

// FastUnmarshalEvent 快速反序列化事件
func (jo *JSONOptimizer) FastUnmarshalEvent(data []byte, event *models.Event) error {
	return json.Unmarshal(data, event)
}

// FastUnmarshalPet 快速反序列化宠物
func (jo *JSONOptimizer) FastUnmarshalPet(data []byte, pet *models.Pet) error {
	return json.Unmarshal(data, pet)
}

// StreamMarshal 流式序列化（用于大量数据）
func (jo *JSONOptimizer) StreamMarshal(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
}

// StreamUnmarshal 流式反序列化
func (jo *JSONOptimizer) StreamUnmarshal(r io.Reader, data interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(data)
}

// BatchMarshalEvents 批量序列化事件
func (jo *JSONOptimizer) BatchMarshalEvents(events []models.Event) ([]byte, error) {
	buf := jo.getBuffer()
	defer jo.putBuffer(buf)
	
	// 手动构建JSON数组以减少内存分配
	buf.WriteByte('[')
	
	for i, event := range events {
		if i > 0 {
			buf.WriteByte(',')
		}
		
		eventData, err := jo.FastMarshalEvent(&event)
		if err != nil {
			return nil, err
		}
		
		// 移除换行符
		eventData = bytes.TrimRight(eventData, "\n")
		buf.Write(eventData)
	}
	
	buf.WriteByte(']')
	
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	
	return result, nil
}

// OptimizedResponse 优化的响应结构
type OptimizedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Count   int         `json:"count,omitempty"`
}

// MarshalResponse 序列化响应
func (jo *JSONOptimizer) MarshalResponse(success bool, data interface{}, errMsg string) ([]byte, error) {
	response := OptimizedResponse{
		Success: success,
		Data:    data,
		Error:   errMsg,
	}
	
	// 如果data是切片，设置count
	if data != nil {
		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Slice {
			response.Count = v.Len()
		}
	}
	
	return jo.FastMarshal(response)
}

// FastMarshal 通用快速序列化
func (jo *JSONOptimizer) FastMarshal(v interface{}) ([]byte, error) {
	buf := jo.getBuffer()
	defer jo.putBuffer(buf)
	
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	
	return result, nil
}

// CacheFieldTags 缓存结构体字段标签
func (jo *JSONOptimizer) CacheFieldTags(structType reflect.Type) {
	if structType.Kind() != reflect.Struct {
		return
	}
	
	if _, exists := jo.tagCache.Load(structType); exists {
		return
	}
	
	fieldMap := make(map[string]string)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			fieldMap[field.Name] = jsonTag
		}
	}
	
	jo.tagCache.Store(structType, fieldMap)
}

// 全局JSON优化器实例
var GlobalJSONOptimizer = NewJSONOptimizer()

// FastMarshalEvent 全局快速事件序列化
func FastMarshalEvent(event *models.Event) ([]byte, error) {
	return GlobalJSONOptimizer.FastMarshalEvent(event)
}

// FastMarshalPet 全局快速宠物序列化
func FastMarshalPet(pet *models.Pet) ([]byte, error) {
	return GlobalJSONOptimizer.FastMarshalPet(pet)
}

// FastMarshalResponse 全局快速响应序列化
func FastMarshalResponse(success bool, data interface{}, errMsg string) ([]byte, error) {
	return GlobalJSONOptimizer.MarshalResponse(success, data, errMsg)
}