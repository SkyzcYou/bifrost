package configuration

import (
	"encoding/json"
	"github.com/ClessLi/bifrost/pkg/resolv/V2/nginx"
	"github.com/ClessLi/bifrost/pkg/resolv/V2/nginx/configuration/parser"
	"github.com/ClessLi/bifrost/pkg/resolv/V2/nginx/dumper"
	"github.com/ClessLi/bifrost/pkg/resolv/V2/nginx/loader"
	"github.com/ClessLi/bifrost/pkg/resolv/V2/nginx/loop_preventer"
	"github.com/ClessLi/bifrost/pkg/resolv/V2/nginx/parser_type"
	"github.com/ClessLi/bifrost/pkg/resolv/V2/utils"
	"sync"
)

//type Updater interface {
//	UpdateFromJsonBytes(data []byte) error
//}

// keyword string: <parser type>[':sep: <value string>', ':sep: :reg: <value regexp>']
//
// e.g. for Nginx Config keyword string:
//     1) server
//     2) location:sep: :reg: \^\~\s+\/
//     3) key:sep: server_name test1\.com
//     4) comment:sep: :reg: .*
type Configuration interface {
	Querier
	// insert
	InsertByKeyword(insertParser parser.Parser, keyword string) error
	InsertByQueryer(insertParser parser.Parser, queryer Querier) error
	//InsertByIndex(insertParser parser.Parser, targetContext parser.Context, index int) error
	// remove
	RemoveByKeyword(keyword string) error
	RemoveByQueryer(queryer Querier) error
	//RemoveByIndex(targetContext parser.Context, index int) error
	// modify
	ModifyByKeyword(modifyParser parser.Parser, keyword string) error
	ModifyByQueryer(modifyParser parser.Parser, queryer Querier) error
	//ModifyByIndex(modifyParser parser.Parser, targetContext parser.Context, index int) error
	// update all
	UpdateFromJsonBytes(data []byte) error

	// view
	View() []byte
	StatisticsByJson() []byte // TODO: 等待割出去
	Json() []byte
	Dump() map[string][]byte

	// private method
	//setConfig(config *parser.Config)
	renewConfiguration(Configuration) error
	//diff(Configuration) bool
	getMainConfigPath() string
	getConfigFingerprinter() utils.ConfigFingerprinter
}

type configuration struct {
	config        *parser.Config
	rwLocker      *sync.RWMutex
	loopPreventer loop_preventer.LoopPreventer
	utils.ConfigFingerprinter
}

func (c *configuration) InsertByKeyword(insertParser parser.Parser, keyword string) error {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()
	parserKeyword, err := parseKeyword(keyword)
	if err != nil {
		return err
	}
	target, idx := c.config.Query(parserKeyword)
	return c.insertByIndex(insertParser, target, idx)
}

func (c *configuration) InsertByQueryer(insertParser parser.Parser, queryer Querier) error {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()
	return c.insertByIndex(insertParser, queryer.fatherContext(), queryer.index())
}

func (c *configuration) insertByIndex(insertParser parser.Parser, targetContext parser.Context, index int) error {
	if insertParser.GetType() == parser_type.TypeConfig {
		err := c.loopPreventer.CheckLoopPrevent(targetContext.GetPosition(), insertParser.GetPosition())
		if err != nil {
			return err
		}
	}
	return targetContext.Insert(insertParser, index)
}

//func (c *configuration) InsertByIndex(insertParser parser.Parser, targetContext parser.Context, index int) error {
//	c.rwLocker.Lock()
//	defer c.rwLocker.Unlock()
//	return targetContext.Insert(insertParser, index)
//}

func (c *configuration) RemoveByKeyword(keyword string) error {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()
	parserKeyword, err := parseKeyword(keyword)
	if err != nil {
		return err
	}
	target, idx := c.config.Query(parserKeyword)
	return target.Remove(idx)
}

func (c *configuration) RemoveByQueryer(queryer Querier) error {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()
	return queryer.fatherContext().Remove(queryer.index())
}

func (c *configuration) removeByIndex(targetContext parser.Context, index int) error {
	if targetContext.GetType() == parser_type.TypeInclude {
		removeParser, err := targetContext.GetChild(index)
		if err != nil {
			return err
		}
		err = c.loopPreventer.RemoveRoute(targetContext.GetPosition(), removeParser.GetPosition())
		if err != nil {
			return err
		}
	}
	return targetContext.Remove(index)
}

//func (c *configuration) RemoveByIndex(targetContext parser.Context, index int) error {
//	c.rwLocker.Lock()
//	defer c.rwLocker.Unlock()
//	return targetContext.Remove(index)
//}

func (c *configuration) ModifyByKeyword(modifyParser parser.Parser, keyword string) error {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()
	parserKeyword, err := parseKeyword(keyword)
	if err != nil {
		return err
	}
	target, idx := c.config.Query(parserKeyword)
	return target.Modify(modifyParser, idx)
}

func (c *configuration) ModifyByQueryer(modifyParser parser.Parser, queryer Querier) error {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()
	return queryer.fatherContext().Modify(modifyParser, queryer.index())
}

//func (c *configuration) ModifyByIndex(modifyParser parser.Parser, targetContext parser.Context, index int) error {
//	c.rwLocker.Lock()
//	defer c.rwLocker.Unlock()
//	return targetContext.Modify(modifyParser, index)
//}

func (c *configuration) UpdateFromJsonBytes(data []byte) error {
	l := loader.NewLoader()
	config, preventer, err := l.LoadFromJsonBytes(data)
	if err != nil {
		return err
	}
	newConfiguration := NewConfiguration(config.(*parser.Config), preventer, new(sync.RWMutex))
	return c.renewConfiguration(newConfiguration)
}

func (c *configuration) Query(keyword string) (Querier, error) {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()
	parserKeyword, err := parseKeyword(keyword)
	if err != nil {
		return nil, err
	}
	foundCtx, index := c.config.Query(parserKeyword)
	if foundCtx == nil {
		return nil, nginx.ErrNotFound
	}
	foundParser, err := foundCtx.GetChild(index)
	if err != nil {
		return nil, err
	}
	return &querier{
		Parser:    foundParser,
		fatherCtx: foundCtx,
		selfIndex: index,
	}, nil
}

func (c *configuration) QueryAll(keyword string) ([]Querier, error) {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()
	parserKeyword, err := parseKeyword(keyword)
	if err != nil {
		return nil, err
	}
	queryers := make([]Querier, 0)
	for foundCtx, indexes := range c.config.QueryAll(parserKeyword) {
		if indexes == nil {
			continue
		}
		for _, index := range indexes {
			foundParser, err := foundCtx.GetChild(index)
			if err != nil {
				return nil, err
			}
			queryers = append(queryers, &querier{
				Parser:    foundParser,
				fatherCtx: foundCtx,
				selfIndex: index,
			})
		}
	}
	return queryers, nil
}

func (c configuration) Self() parser.Parser {
	return c.config
}

func (c configuration) fatherContext() parser.Context {
	return nil
}

func (c configuration) index() int {
	return 0
}

func (c *configuration) View() []byte {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()
	return c.config.Bytes()
}

func (c *configuration) Json() []byte {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()
	data, err := json.Marshal(c.config)
	if err != nil {
		return nil
	}
	return data
}

func (c *configuration) StatisticsByJson() []byte {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()
	statistician := NewStatistician(c)
	statistics := statistician.Statistics()
	data, err := json.Marshal(statistics)
	if err != nil {
		return nil
	}
	return data
}

func (c *configuration) Dump() map[string][]byte {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()
	d := dumper.NewDumper(c.config.GetValue())
	_ = c.config.Dump(d)
	return d.ReadAll()
}

func (c *configuration) renewConfiguration(conf Configuration) error {
	if !c.ConfigFingerprinter.Diff(conf.getConfigFingerprinter()) {
		return ErrSameConfigFingerprint
	}
	newConf, ok := conf.(*configuration)
	if !ok {
		return ErrConfigurationTypeMismatch
	}
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()
	c.config = newConf.config
	c.ConfigFingerprinter.Renew(newConf.ConfigFingerprinter)
	c.loopPreventer = newConf.loopPreventer
	return nil
}

func (c *configuration) getMainConfigPath() string {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()
	return c.config.GetValue()
}

func (c *configuration) getConfigFingerprinter() utils.ConfigFingerprinter {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()
	return c.ConfigFingerprinter
}

func NewConfiguration(config *parser.Config, preventer loop_preventer.LoopPreventer, rwLocker *sync.RWMutex) Configuration {
	conf := &configuration{
		rwLocker:      rwLocker,
		loopPreventer: preventer,
		config:        config,
	}
	conf.ConfigFingerprinter = utils.NewConfigFingerprinter(conf.Dump())
	return conf
}
