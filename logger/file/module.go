package filestorage

import (
    "os"
    "sync"
)

type FileStorageConf struct {
    Directory string
    FSync bool
}

type FileStorage struct {
    file *os.File
    Conf FileStorageConf
    mu sync.Mutex
}

func NewFileStorage() *FileStorage {
    storage := &FileStorage{}
    return storage
}

func (p *FileStorage) InitModule() error {
    filename := p.Conf.Directory + "log.requests"
    var err error
    p.file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }

    return nil
}

func (p *FileStorage) WriteLine(s string) error {
    p.mu.Lock()
    defer p.mu.Unlock()

    _, err := p.file.WriteString(s)
    if err != nil {
        return err
    }

    if p.Conf.FSync {
        err = p.file.Sync()
    }
    return err
}
