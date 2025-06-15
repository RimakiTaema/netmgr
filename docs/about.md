<script setup>
import { VPTeamMembers } from 'vitepress/theme'

const members = [
  {
    avatar: 'https://github.com/RimakiTaema.png',
    name: 'Rimaki',
    title: 'Lead Dev & Maintainer',
    links: [
      { icon: 'github', link: 'https://github.com/RimakiTaema' },
      { icon: 'link', link: 'https://rimakiproject.online' }
    ]
  }
]
</script>

# About netmgr

Netmgr is an enterprise-grade, cross-platform network control and wrapper utility built in **C++ & C**. Designed for system administrators, developers, and advanced users, it directly leverages native tools available on each platform—without relying on abstractions like `firewalld` or `ufw`.

Its goal is to provide **precise control**, automation capabilities, and a future-proof CLI and scripting interface across Linux, Windows, and macOS.

---

## Supported Platforms

| OS        | Primary Tools Used                            |
|-----------|----------------------------------------------|
| **Linux** | `ip`, `iptables`, `tc`                        |
| **Windows** | `netsh`, `PowerShell`                        |
| **macOS** | `networksetup`, `scutil`, `pfctl`             |

No dependency on tools like `firewalld`, `ufw`, or GUI wrappers—`netmgr` speaks directly to the system.

---

## Key Features

- **Advanced IPTables Integration**  
  Configure complex routing, NAT, MTU, and port filtering in a structured and repeatable way.

- **Cross-Platform Support**  
  Native wrappers and behaviors per OS without breaking conventions. Windows supports both `TCP` and `UDP`.

- **Enterprise Ready**  
  Ships with optional build profiles, licensing controls, and internal audit logs (planned).

- **Script & Automation Friendly**  
  Designed for CI/CD pipelines, network testing, failover setups, or rapid config reloads.

---

## Why Not `ufw` or `firewalld`?

`netmgr` is built to offer **fine-grained control** using native system-level tools, ideal for:

- Security-focused deployments
- Debugging or testing low-level configurations
- Managing advanced networking scenarios (VPNs, tunnels, NAT configs, etc.)
- Reducing abstraction and latency in packet handling

---

## License

- **Open Source Edition:** MIT License  
- **Enterprise Edition:** Requires a commercial license. See [Enterprise License Guide](/guides/enterprise/licence.md)

---

## Maintainer

<VPTeamMembers size="medium" :members="members" />

---

## Screenshots & Usage

🚧 *Coming soon as part of the `examples/` section.*

---

## Contributions

All contributions are welcome! Open an issue, fork the repo, or join the development discussion via GitHub.

---

> _“netmgr was built for people who need full control—not convenience with limits.”_
