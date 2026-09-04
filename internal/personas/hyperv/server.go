// Copyright 2026 Zyvor AI Labs
// SPDX-License-Identifier: Apache-2.0

package hyperv

import (
	"crypto/subtle"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/zyvorai/chimera/internal/personas/common"
)

type Server struct {
	Username, Password string
	Store              *common.Store
}

func New(username, password string, vmCount int) *Server {
	return &Server{Username: username, Password: password, Store: common.NewStore(vmCount)}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.auth(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="WSMan"`)
		http.Error(w, "unauthorized", 401)
		return
	}
	if r.URL.Path != "/wsman" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var env envelope
	if err := xml.NewDecoder(r.Body).Decode(&env); err != nil {
		soapFault(w, "InvalidEnvelope", err.Error())
		return
	}
	action := strings.ToLower(env.Header.Action)
	switch {
	case strings.Contains(action, "identify"):
		s.identify(w)
	case strings.Contains(action, "enumerate"):
		s.enumerate(w)
	case strings.Contains(action, "pull"):
		s.pull(w)
	case strings.Contains(action, "invoke") || strings.Contains(action, "requeststatechange"):
		s.power(w, env)
	default:
		soapFault(w, "ActionNotSupported", env.Header.Action)
	}
}

type envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Header  struct {
		Action      string `xml:"Action"`
		ResourceURI string `xml:"ResourceURI"`
		SelectorSet struct {
			Selectors []struct {
				Name  string `xml:"Name,attr"`
				Value string `xml:",chardata"`
			} `xml:"Selector"`
		} `xml:"SelectorSet"`
	} `xml:"Header"`
	Body struct {
		XML string `xml:",innerxml"`
	} `xml:"Body"`
}

func (s *Server) identify(w http.ResponseWriter) {
	soap(w, `<wsmid:IdentifyResponse xmlns:wsmid="http://schemas.dmtf.org/wbem/wsman/identity/1/wsmanidentity.xsd"><wsmid:ProtocolVersion>http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd</wsmid:ProtocolVersion><wsmid:ProductVendor>Zyvor Chimera</wsmid:ProductVendor><wsmid:ProductVersion>Hyper-V 2025 compatible</wsmid:ProductVersion></wsmid:IdentifyResponse>`)
}
func (s *Server) enumerate(w http.ResponseWriter) {
	soap(w, `<wsen:EnumerateResponse xmlns:wsen="http://schemas.xmlsoap.org/ws/2004/09/enumeration"><wsen:EnumerationContext>chimera-vms</wsen:EnumerationContext></wsen:EnumerateResponse>`)
}
func (s *Server) pull(w http.ResponseWriter) {
	var b strings.Builder
	b.WriteString(`<wsen:PullResponse xmlns:wsen="http://schemas.xmlsoap.org/ws/2004/09/enumeration"><wsen:Items>`)
	for _, vm := range s.Store.List() {
		fmt.Fprintf(&b, `<p:Msvm_ComputerSystem xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wmi/root/virtualization/v2/Msvm_ComputerSystem"><p:Name>%s</p:Name><p:ElementName>%s</p:ElementName><p:EnabledState>%d</p:EnabledState><p:Description>Microsoft Virtual Machine</p:Description></p:Msvm_ComputerSystem>`, vm.ID, xmlEscape(vm.Name), enabledState(vm.Power))
	}
	b.WriteString(`</wsen:Items><wsen:EndOfSequence/></wsen:PullResponse>`)
	soap(w, b.String())
}
func (s *Server) power(w http.ResponseWriter, env envelope) {
	id := ""
	for _, sel := range env.Header.SelectorSet.Selectors {
		if strings.EqualFold(sel.Name, "Name") {
			id = sel.Value
		}
	}
	if id == "" {
		soapFault(w, "InvalidSelectors", "Name selector required")
		return
	}
	state := "ON"
	if strings.Contains(env.Body.XML, ">3<") || strings.Contains(strings.ToLower(env.Body.XML), "off") {
		state = "OFF"
	}
	t, ok := s.Store.SetPower(id, state)
	if !ok {
		soapFault(w, "NotFound", id)
		return
	}
	soap(w, fmt.Sprintf(`<p:RequestStateChange_OUTPUT xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wmi/root/virtualization/v2/Msvm_ComputerSystem"><p:ReturnValue>4096</p:ReturnValue><p:Job>%s</p:Job></p:RequestStateChange_OUTPUT>`, t.ID))
}

func (s *Server) auth(r *http.Request) bool {
	u, p, ok := r.BasicAuth()
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(u), []byte(s.Username)) == 1 && subtle.ConstantTimeCompare([]byte(p), []byte(s.Password)) == 1
}
func enabledState(p string) int {
	if p == "ON" {
		return 2
	}
	return 3
}
func soap(w http.ResponseWriter, inner string) {
	w.Header().Set("Content-Type", "application/soap+xml;charset=UTF-8")
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?><s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>%s</s:Body></s:Envelope>`, inner)
}
func soapFault(w http.ResponseWriter, code, msg string) {
	w.WriteHeader(500)
	soap(w, fmt.Sprintf(`<s:Fault xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Code><s:Value>s:Receiver</s:Value></s:Code><s:Reason><s:Text>%s: %s</s:Text></s:Reason></s:Fault>`, xmlEscape(code), xmlEscape(msg)))
}
func xmlEscape(v string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(v))
	return b.String()
}
