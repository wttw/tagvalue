package tagvalue

import (
	"fmt"
	"testing"
)

var sig1 = " v=1; a=rsa-sha256; d=example.net; s=brisbane;\r\n" +
" c=simple; q=dns/txt; i=@eng.example.net;\r\n" +
" t=1117574938; x=1118006938;\r\n" +
" h=from:to:subject:date;\r\n" +
" z=From:foo@eng.example.net|To:joe@example.com|\r\n" +
" Subject:demo=20run|Date:July=205,=202005=203:44:08=20PM=20-0700;\r\n" +
" bh=MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=;\r\n" +
" b=dzdVyOfAKCdLXdJOc9G2q8LoXSlEniSbav+yuU4zGeeruD00lszZVoG4ZHRNiYzR"

var sig2 = "v=1; a=rsa-sha256; s=brisbane; d=example.com;\t\r\n" +
" \t c=simple/simple; q=dns/txt; i=joe@football.example.com;\r\n" +
" h=Received : From : To : Subject : Date : Message-ID;\r\n" +
" bh=2jUSOH9NhtVGCQWNr9BrIAPreKQjO6Sn7XIkfJVOzv8=;\r\n" +
" b=AuUoFEfDxTDkHlLXSZEpZj79LICEps6eda7W3deTVFOk4yAUoqOB\r\n" +
" 4nujc7YopdG5dWLSdNg6xNAZpOPr+kHxt1IrE+NahM6L/LbvaHut\r\n" +
" KVdkLLkpVaVVQPzeRDI009SO2Il5Lu7rDNH6mZckBdrIx0orEtZV\r\n" +
" 4bmp/YzhwvcubU4=;"

var key1 = "v=DKIM1; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQ\r\n" +
" KBgQDwIRP/UC3SBsEmGqZ9ZJW3/DkMoGeLnQg1fWn7/zYt\r\n" +
" IxN2SnFCjxOCKG9v3b4jYfcTNh5ijSsq631uBItLa7od+v\r\n" +
" /RtdC2UzJ1lWT947qR+Rcac2gbto/NMqJ0fzfVjH4OuKhi\r\n" +
" tdY9tf6mcwGjaNBcWToIMmPSPDdQPNUYckcQ2QIDAQAB"

func TestNewTagValue(t *testing.T) {
	v, err := NewTagValue(sig1)
	if err != nil {
		t.Errorf("sig1 parse failed: %v", err)
	}
	fmt.Printf("v=%#v\n", v)
}
