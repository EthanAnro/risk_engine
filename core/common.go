package core

import (
	"errors"
	"fmt"
	"github.com/skyhackvip/risk_engine/internal/operator"
	"log"
)

type NodeInfo struct {
	Id      int64    `yaml:"id"`
	Name    string   `yaml:"name"`
	Tag     string   `yaml:"tag"`
	Label   string   `yaml:"label"`
	Kind    string   `yaml:"kind"`
	Depends []string `yaml:"depends,flow"`
}

type BlockStrategy struct {
	IsBock   bool        `yaml:"is_block"`
	HitRule  []string    `yaml:"hit_rule,flow"`
	Operator string      `yaml:"opeartor"`
	Value    interface{} `yaml:"value"`
}

type Rule struct {
	Id         string      `yaml:"id"`
	Name       string      `yaml:"name"`
	Tag        string      `yaml:"tag"`
	Label      string      `yaml:"label"`
	Conditions []Condition `yaml:"conditions,flow"`
	Decision   Decision    `yaml:"decision"`
}

//parse rule
func (rule *Rule) Parse(ctx *PipelineContext) (output *Output, err error) {
	//rule.Conditions
	if len(rule.Conditions) == 0 {
		err = errors.New(fmt.Sprintf("rule (%s) condition is empty", rule.Name))
		return
	}
	var conditionRet = make(map[string]interface{}, 0)
	for _, condition := range rule.Conditions {
		if feature, ok := ctx.GetFeature(condition.Feature); ok {
			value, _ := feature.GetValue() //是否使用default
			rs, err := operator.Compare(condition.Operator, value, condition.Value)
			if err != nil {
				return output, nil
			}
			conditionRet[condition.Name] = rs
		} else {
			//lack of feature  whether ignore
			log.Printf("error lack of feature: %s\n", condition.Feature)
			continue
		}
	}
	if len(conditionRet) == 0 {
		err = errors.New(fmt.Sprintf("rule (%s) condition result error", rule.Name))
		return
	}

	//rule.Decision
	expr := rule.Decision.Logic
	logicRet, err := operator.Evaluate(expr, conditionRet)
	if err != nil {
		return
	}
	log.Printf("rule %s (%s) decision is: %v, output: %v\n", rule.Label, rule.Name, logicRet, rule.Decision.Output)

	//assign
	if len(rule.Decision.Assign) > 0 {
		features := make(map[string]*Feature)
		for name, value := range rule.Decision.Assign {
			feature := NewFeature(name, TypeString, "") //string
			feature.SetValue(value)
			features[name] = feature
		}
		ctx.SetFeatures(features)
	}
	return &rule.Decision.Output, nil
}

type Condition struct {
	Feature  string      `yaml:"feature"`
	Operator string      `yaml:"operator"`
	Value    interface{} `yaml:"value"`
	Result   string      `yaml:"result"`
	Name     string      `yaml:"name"`
}

type Decision struct {
	Depends []string               `yaml:"depends,flow"` //依赖condition结果
	Logic   string                 `yaml:"logic"`
	Output  Output                 `yaml:"output"`
	Assign  map[string]interface{} `yaml:"assign"` //赋值更多变量
}

type Output struct {
	Name  string      `yaml:"name"` //该节点输出值重命名，如果无则以（节点类型+节点名）赋值变量
	Value interface{} `yaml:"value"`
	Kind  string      `yaml:"kind"` //nodetype featuretype
}

type Branch struct {
	Name       string      `yaml:"name"`
	Conditions []Condition `yaml:"conditions"` //used by conditional
	Logic      string      `yaml:"logic"`      //used by conditional
	Percent    float64     `yaml:"percent"`    //used by abtest
	Decision   Decision    `yaml:"decision"`
}