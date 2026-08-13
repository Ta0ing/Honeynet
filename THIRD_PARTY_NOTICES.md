# Third-Party Notices

Honeynet includes or references third-party software and imported data. The repository-level noncommercial license applies only to original Honeynet code. It does not replace, restrict, or grant rights under any third-party license.

Copyright, trademark, patent, attribution, source-disclosure, and redistribution obligations imposed by third-party licenses remain in effect. Distributors of source or binaries must retain the complete applicable upstream license texts, copyright notices, and `NOTICE` files. This inventory is not a substitute for those texts.

## Go direct dependencies

Versions below are taken from `go.mod`.

| Component | Version | License |
|---|---:|---|
| `github.com/ClickHouse/clickhouse-go/v2` | 2.40.3 | Apache-2.0 |
| `github.com/gin-contrib/cors` | 1.7.6 | MIT |
| `github.com/gin-gonic/gin` | 1.10.1 | MIT |
| `github.com/golang-jwt/jwt/v5` | 5.2.2 | MIT |
| `github.com/google/uuid` | 1.6.0 | BSD-3-Clause |
| `github.com/gorilla/websocket` | 1.5.3 | BSD-2-Clause |
| `github.com/ipipdotnet/ipdb-go` | 1.3.3 | Apache-2.0 |
| `golang.org/x/crypto` | 0.52.0 | BSD-3-Clause |
| `golang.org/x/image` | 0.45.0 | BSD-3-Clause |
| `golang.org/x/sys` | 0.47.0 | BSD-3-Clause |
| `gopkg.in/yaml.v3` | 3.0.1 | MIT AND Apache-2.0; includes Canonical Ltd. NOTICE |
| `gorm.io/datatypes` | 1.2.7 | MIT |
| `gorm.io/driver/mysql` | 1.6.0 | MIT |
| `gorm.io/driver/sqlite` | 1.6.0 | MIT |
| `gorm.io/gorm` | 1.30.1 | MIT |

Transitive versions and their integrity hashes are recorded in `go.sum`. Their upstream license texts and notices remain controlling.

## Web direct dependencies

Exact installed versions below are taken from `web/package-lock.json`.

| Component | Version | License |
|---|---:|---|
| `@arco-design/web-react` | 2.66.16 | MIT |
| `@tanstack/react-query` | 5.101.4 | MIT |
| `axios` | 1.19.0 | MIT |
| `dayjs` | 1.11.21 | MIT |
| `react` | 18.3.1 | MIT |
| `react-dom` | 18.3.1 | MIT |
| `react-router-dom` | 7.18.2 | MIT |
| `zustand` | 5.0.14 | MIT |
| `@types/react` | 18.3.31 | MIT |
| `@types/react-dom` | 18.3.7 | MIT |
| `@vitejs/plugin-react` | 4.7.0 | MIT |
| `typescript` | 5.8.3 | Apache-2.0 |
| `vite` | 6.4.3 | MIT |

The lockfile also contains transitive packages under MIT, Apache-2.0, BSD, ISC, CC-BY-4.0, and other licenses. A distributor must generate and retain the complete notice inventory for the exact lockfile used to build a release.

## Honeypot template snapshots

`honeypot-templates-server/services/` is an imported collection of product-page snapshots and static assets. It has no directory-wide license, notice, provenance manifest, or redistribution grant in this repository.

Identifiable embedded declarations include, without limitation:

- Apache-2.0 material, including Apache Tomcat files and OpenWrt LuCI resources.
- GPL material, including Joomla, Poste, Zabbix, Zimbra, and JXON files. One retained jQuery EasyUI file states only “Licensed under the GPL terms” without identifying a GPL version.
- MIT and dual MIT/GPL components such as jQuery-related plugins, Bootstrap, Vue, classnames, normalize.css, and RequireJS.
- Material marked “All Rights Reserved”, “VMware Confidential”, “Licensed Materials - Property of IBM”, or subject to an Atlassian end-user agreement. Similar copyright assertions appear in Microsoft, Intel, Hikvision, Synology, Oracle, Nagios, NSFOCUS, Sangfor, and other product snapshots.

The Honeynet license does not apply to or relicense these files. A notice file cannot create missing redistribution permission. Anyone redistributing this directory or a release containing it is responsible for establishing the right to redistribute every included file and satisfying all applicable license-text, notice, source-offer, and attribution obligations. Remove material for which those rights cannot be established.

## Decrypted CVE rule collection

`cve-rules-decrypted/` contains imported JSON template and YARA rule files. The retained collection has no directory-wide README, license, copyright notice, author field, source URL, or provenance record establishing ownership or redistribution permission.

These files are not relicensed by the Honeynet license. Anyone redistributing them is responsible for documenting their authorship, source, applicable license, modification history, and redistribution permission. Remove them if those rights cannot be established.

## Trademarks

Product names, logos, screenshots, and trade dress remain the property of their respective owners. Their presence does not imply affiliation, sponsorship, or endorsement.
