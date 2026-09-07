package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"time"

	bolt "go.etcd.io/bbolt"
)

// ── 加密工具 ──────────────────────────────────────────────────────────────────

var masterKey []byte

func SetMasterKey(key string) {
	h := sha256.Sum256([]byte(key))
	masterKey = h[:]
}

func Encrypt(plaintext string) (string, error) {
	if len(masterKey) == 0 {
		return "", errors.New("master key not set")
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(encoded string) (string, error) {
	if len(masterKey) == 0 {
		return "", errors.New("master key not set")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// ── 数据模型 ──────────────────────────────────────────────────────────────────

type User struct {
	ID           uint      `json:"id"`
	Username     string    `json:"username"`
	Password     string    `json:"-"`
	TokenVersion int       `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Settings 系统设置。
// JWTSecretEnc      — AES-GCM 加密的 JWT 签名密钥（base64）
// WebDAVPasswordEnc — AES-GCM 加密的 WebDAV 独立密码 bcrypt 哈希（base64）
type Settings struct {
	ID                uint   `json:"id"`
	StorageDir        string `json:"storage_dir"`
	Lang              string `json:"lang"`
	UITheme           string `json:"ui_theme"`
	UIFont            string `json:"ui_font"`
	EditorFont        string `json:"editor_font"`
	WebDAVEnabled     bool   `json:"webdav_enabled"`
	WebDAVSubPath     string `json:"webdav_sub_path"`
	WebDAVUsername    string `json:"webdav_username"`
	WebDAVPasswordEnc string `json:"-"` // AES-GCM(bcrypt(password))
	JWTSecretEnc      string `json:"-"` // AES-GCM(jwt_secret)
	ShowHidden        bool   `json:"show_hidden"`
	FileViewMode      string `json:"file_view_mode"`  // list | detail | icon-large | icon-small
	FileSortBy        string `json:"file_sort_by"`    // name | date | type | size
	FileSortOrder     string `json:"file_sort_order"` // asc | desc
	GithubProxyEnabled bool  `json:"github_proxy_enabled"`
}

func (s *Settings) GetJWTSecret() (string, error) {
	if s.JWTSecretEnc == "" {
		return "", nil
	}
	return Decrypt(s.JWTSecretEnc)
}

func (s *Settings) SetJWTSecret(secret string) error {
	enc, err := Encrypt(secret)
	if err != nil {
		return err
	}
	s.JWTSecretEnc = enc
	return nil
}

func (s *Settings) GetWebDAVPasswordHash() (string, error) {
	if s.WebDAVPasswordEnc == "" {
		return "", nil
	}
	return Decrypt(s.WebDAVPasswordEnc)
}

func (s *Settings) SetWebDAVPasswordHash(bcryptHash string) error {
	if bcryptHash == "" {
		s.WebDAVPasswordEnc = ""
		return nil
	}
	enc, err := Encrypt(bcryptHash)
	if err != nil {
		return err
	}
	s.WebDAVPasswordEnc = enc
	return nil
}

type ShareLink struct {
	ID        uint       `json:"id"`
	Code      string     `json:"code"`
	FilePath  string     `json:"file_path"`
	IsDir     bool       `json:"is_dir"`
	UserID    uint       `json:"user_id"`
	ExpiresAt *time.Time `json:"expires_at"`
	MaxViews  int        `json:"max_views"`
	ViewCount int        `json:"view_count"`
	CreatedAt time.Time  `json:"created_at"`
}

type FileVisibility struct {
	ID       uint   `json:"id"`
	FilePath string `json:"file_path"`
	IsPublic bool   `json:"is_public"`
}

// ── bbolt bucket 名称 ─────────────────────────────────────────────────────────

var (
	bucketUsers      = []byte("users")
	bucketSettings   = []byte("settings")
	bucketShares     = []byte("shares")
	bucketShareCode  = []byte("share_codes")  // code -> id 索引
	bucketVisibility = []byte("visibility")
	bucketSeq        = []byte("sequences")
)

// ── DB 封装 ───────────────────────────────────────────────────────────────────

// DB 是对 bbolt 数据库的封装，提供与原 gorm.DB 等价的操作接口。
type DB struct {
	bolt *bolt.DB
}

// itob 将 uint64 转换为大端序 8 字节（用作 bbolt key）
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func btoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// nextSeq 返回指定 bucket 的下一个自增 ID
func (d *DB) nextSeq(tx *bolt.Tx, name string) (uint64, error) {
	bkt := tx.Bucket(bucketSeq)
	if bkt == nil {
		return 0, errors.New("sequences bucket missing")
	}
	key := []byte(name)
	cur := uint64(0)
	if v := bkt.Get(key); v != nil {
		cur = btoi(v)
	}
	next := cur + 1
	return next, bkt.Put(key, itob(next))
}

// ── 数据库初始化 ───────────────────────────────────────────────────────────────

func InitDB(path string) (*DB, error) {
	bdb, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, err
	}

	// 创建所有 bucket
	err = bdb.Update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{
			bucketUsers, bucketSettings, bucketShares,
			bucketShareCode, bucketVisibility, bucketSeq,
		} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	db := &DB{bolt: bdb}

	// 若没有 settings 记录，写入默认值
	s, _ := db.GetSettings()
	if s == nil {
		def := &Settings{
			StorageDir:   "./data/storage",
			Lang:         "zh",
			FileViewMode: "detail",
			FileSortBy:   "name",
			FileSortOrder: "asc",
		}
		_ = db.SaveSettings(def)
	}

	return db, nil
}

// Close 关闭底层 bbolt 数据库
func (d *DB) Close() error {
	return d.bolt.Close()
}

// ── User CRUD ─────────────────────────────────────────────────────────────────

func (d *DB) CreateUser(u *User) error {
	return d.bolt.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketUsers)
		id, err := d.nextSeq(tx, "users")
		if err != nil {
			return err
		}
		u.ID = uint(id)
		u.CreatedAt = time.Now()
		u.UpdatedAt = time.Now()
		data, err := json.Marshal(u)
		if err != nil {
			return err
		}
		return bkt.Put(itob(id), data)
	})
}

func (d *DB) GetUserByUsername(username string) (*User, error) {
	var found *User
	err := d.bolt.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketUsers)
		return bkt.ForEach(func(k, v []byte) error {
			var u User
			if err := json.Unmarshal(v, &u); err != nil {
				return nil
			}
			if u.Username == username {
				found = &u
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, errors.New("user not found")
	}
	return found, nil
}

func (d *DB) GetUserByID(id uint) (*User, error) {
	var u User
	err := d.bolt.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketUsers)
		v := bkt.Get(itob(uint64(id)))
		if v == nil {
			return errors.New("user not found")
		}
		return json.Unmarshal(v, &u)
	})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (d *DB) SaveUser(u *User) error {
	return d.bolt.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketUsers)
		u.UpdatedAt = time.Now()
		data, err := json.Marshal(u)
		if err != nil {
			return err
		}
		return bkt.Put(itob(uint64(u.ID)), data)
	})
}

func (d *DB) CountUsers() (int64, error) {
	var count int64
	err := d.bolt.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketUsers)
		count = int64(bkt.Stats().KeyN)
		return nil
	})
	return count, err
}

// GetFirstUser 返回 ID 最小的用户（用于 WebDAV 单用户场景）
func (d *DB) GetFirstUser() (*User, error) {
	var found *User
	err := d.bolt.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketUsers)
		cur := bkt.Cursor()
		k, v := cur.First()
		if k == nil {
			return errors.New("no users")
		}
		var u User
		if err := json.Unmarshal(v, &u); err != nil {
			return err
		}
		found = &u
		return nil
	})
	return found, err
}

// ── Settings CRUD ─────────────────────────────────────────────────────────────

var settingsKey = itob(1)

func (d *DB) GetSettings() (*Settings, error) {
	var s Settings
	err := d.bolt.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketSettings)
		v := bkt.Get(settingsKey)
		if v == nil {
			return errors.New("no settings")
		}
		return json.Unmarshal(v, &s)
	})
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (d *DB) SaveSettings(s *Settings) error {
	s.ID = 1
	return d.bolt.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketSettings)
		data, err := json.Marshal(s)
		if err != nil {
			return err
		}
		return bkt.Put(settingsKey, data)
	})
}

// ── ShareLink CRUD ────────────────────────────────────────────────────────────

func (d *DB) CreateShare(link *ShareLink) error {
	return d.bolt.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketShares)
		idx := tx.Bucket(bucketShareCode)
		id, err := d.nextSeq(tx, "shares")
		if err != nil {
			return err
		}
		link.ID = uint(id)
		link.CreatedAt = time.Now()
		data, err := json.Marshal(link)
		if err != nil {
			return err
		}
		if err := bkt.Put(itob(id), data); err != nil {
			return err
		}
		return idx.Put([]byte(link.Code), itob(id))
	})
}

func (d *DB) GetShareByCode(code string) (*ShareLink, error) {
	var link ShareLink
	err := d.bolt.View(func(tx *bolt.Tx) error {
		idx := tx.Bucket(bucketShareCode)
		idBytes := idx.Get([]byte(code))
		if idBytes == nil {
			return errors.New("share not found")
		}
		bkt := tx.Bucket(bucketShares)
		v := bkt.Get(idBytes)
		if v == nil {
			return errors.New("share not found")
		}
		return json.Unmarshal(v, &link)
	})
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (d *DB) UpdateShareViewCount(link *ShareLink) error {
	return d.bolt.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketShares)
		data, err := json.Marshal(link)
		if err != nil {
			return err
		}
		return bkt.Put(itob(uint64(link.ID)), data)
	})
}

func (d *DB) ListSharesByUser(userID uint) ([]ShareLink, error) {
	var links []ShareLink
	err := d.bolt.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketShares)
		return bkt.ForEach(func(k, v []byte) error {
			var l ShareLink
			if err := json.Unmarshal(v, &l); err != nil {
				return nil
			}
			if l.UserID == userID {
				links = append(links, l)
			}
			return nil
		})
	})
	return links, err
}

func (d *DB) DeleteShare(id uint, userID uint) error {
	return d.bolt.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketShares)
		idx := tx.Bucket(bucketShareCode)
		v := bkt.Get(itob(uint64(id)))
		if v == nil {
			return errors.New("share not found")
		}
		var l ShareLink
		if err := json.Unmarshal(v, &l); err != nil {
			return err
		}
		if l.UserID != userID {
			return errors.New("permission denied")
		}
		_ = idx.Delete([]byte(l.Code))
		return bkt.Delete(itob(uint64(id)))
	})
}

// ── FileVisibility CRUD ───────────────────────────────────────────────────────

func (d *DB) GetVisibility(filePath string) (*FileVisibility, error) {
	var vis FileVisibility
	err := d.bolt.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketVisibility)
		v := bkt.Get([]byte(filePath))
		if v == nil {
			return errors.New("not found")
		}
		return json.Unmarshal(v, &vis)
	})
	if err != nil {
		return nil, err
	}
	return &vis, nil
}

func (d *DB) SetVisibility(filePath string, isPublic bool) error {
	return d.bolt.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketVisibility)
		vis := FileVisibility{FilePath: filePath, IsPublic: isPublic}
		data, err := json.Marshal(vis)
		if err != nil {
			return err
		}
		return bkt.Put([]byte(filePath), data)
	})
}

// DeleteVisibility 删除文件/目录及其所有子路径的可见性记录
func (d *DB) DeleteVisibilityPrefix(prefix string) error {
	return d.bolt.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketVisibility)
		var toDelete [][]byte
		_ = bkt.ForEach(func(k, _ []byte) error {
			key := string(k)
			if key == prefix || len(key) > len(prefix) && key[:len(prefix)] == prefix && key[len(prefix)] == '/' {
				toDelete = append(toDelete, append([]byte{}, k...))
			}
			return nil
		})
		for _, k := range toDelete {
			if err := bkt.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
}

// RenameVisibilityPrefix 将所有以 src 为前缀的可见性记录重命名为 dst 前缀
func (d *DB) RenameVisibilityPrefix(src, dst string) error {
	return d.bolt.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketVisibility)
		type kv struct{ k, v []byte }
		var matched []kv
		_ = bkt.ForEach(func(k, v []byte) error {
			key := string(k)
			if key == src || len(key) > len(src) && key[:len(src)] == src && key[len(src)] == '/' {
				matched = append(matched, kv{append([]byte{}, k...), append([]byte{}, v...)})
			}
			return nil
		})
		for _, m := range matched {
			oldKey := string(m.k)
			newKey := dst + oldKey[len(src):]
			// 更新值中的 file_path
			var vis FileVisibility
			if err := json.Unmarshal(m.v, &vis); err == nil {
				vis.FilePath = newKey
				if data, err := json.Marshal(vis); err == nil {
					m.v = data
				}
			}
			_ = bkt.Delete(m.k)
			_ = bkt.Put([]byte(newKey), m.v)
		}
		return nil
	})
}

// ListPublicVisibility 返回所有公开的文件路径
func (d *DB) ListPublicVisibility() ([]FileVisibility, error) {
	var result []FileVisibility
	err := d.bolt.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketVisibility)
		return bkt.ForEach(func(k, v []byte) error {
			var vis FileVisibility
			if err := json.Unmarshal(v, &vis); err != nil {
				return nil
			}
			if vis.IsPublic {
				result = append(result, vis)
			}
			return nil
		})
	})
	return result, err
}

// ListVisibilityByPaths 批量查询指定路径的可见性
func (d *DB) ListVisibilityByPaths(paths []string) ([]FileVisibility, error) {
	pathSet := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		pathSet[p] = struct{}{}
	}
	var result []FileVisibility
	err := d.bolt.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketVisibility)
		for _, p := range paths {
			v := bkt.Get([]byte(p))
			if v == nil {
				continue
			}
			var vis FileVisibility
			if err := json.Unmarshal(v, &vis); err == nil && vis.IsPublic {
				result = append(result, vis)
			}
		}
		return nil
	})
	return result, err
}

// ListPublicVisibilityByPrefix 返回指定路径前缀下所有公开的记录
func (d *DB) ListPublicVisibilityByPrefix(prefix string) ([]FileVisibility, error) {
	var result []FileVisibility
	err := d.bolt.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketVisibility)
		return bkt.ForEach(func(k, v []byte) error {
			key := string(k)
			if key != prefix && !(len(key) > len(prefix) && key[:len(prefix)] == prefix && key[len(prefix)] == '/') {
				return nil
			}
			var vis FileVisibility
			if err := json.Unmarshal(v, &vis); err == nil && vis.IsPublic {
				result = append(result, vis)
			}
			return nil
		})
	})
	return result, err
}
