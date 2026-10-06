# 🚀 Como Instalar e Usar o MoveOps (Guia Rápido e Descomplicado)

Seja muito bem-vindo ao **MoveOps**!  
Este guia foi feito para ser simples, direto ao ponto e acessível para **qualquer pessoa**, mesmo que você nunca tenha mexido em servidores ou linhas de comando.

---

## 🎯 O que é o MoveOps e Como Ele Funciona?

O **MoveOps** é uma ferramenta profissional e ultrarrápida para cópia e sincronização de grandes volumes de arquivos e pastas.

- 💻 **Ele roda direto no seu computador** (no Windows ou no Linux).
- 🌐 **Você controla tudo pelo seu navegador de internet favorito** (Google Chrome, Edge, Firefox, Brave, Safari).
- 🛡️ **Segurança Total**: Ele já vem configurado com o modo **"Produção Segura"**, o que significa que ele não vai travar o seu computador, não vai congestionar sua rede e não vai deixar o seu disco lento para outras tarefas.
- ⚡ **Sem Instalar Nada Extra**: Não precisa instalar Java, Python, Node.js ou bancos de dados. É baixar e usar!

---

## 🪟 Como Instalar e Rodar no Windows

Escolha a forma que for mais fácil para você:

### Método 1: Recomendado — Instalador em 1 Clique (Instalação Tradicional)
Ideal se você quer que o MoveOps fique instalado como um programa comum no seu Windows.

1. **Baixe ou localize** o arquivo:
   ```
   MoveOps-Setup.exe
   ```
   *(Ou execute com 1 clique o arquivo `Instalar-MoveOps.bat` presente na pasta do projeto).*
2. **Dê duplo clique** no instalador.
3. Clique em **"Avançar"**, **"Avançar"** e depois em **"Concluir"**.
4. ✨ **Pronto!** O instalador faz tudo para você:
   - Cria o ícone oficial do **MoveOps** na sua **Área de Trabalho** e no **Menu Iniciar**.
   - Configura as permissões do **Firewall do Windows** automaticamente (para você não precisar configurar nada manualmente).
   - Inicia o MoveOps e abre seu navegador na tela inicial!

---

### Método 2: Super Rápido / Portátil (Sem Precisar Instalar Nada)
Ideal se você quer apenas testar rapidamente ou rodar direto de um pendrive/pasta de rede.

1. Abra a pasta do MoveOps no seu computador.
2. Dê um **duplo clique** no arquivo:
   ```
   Iniciar-MoveOps.bat
   ```
   *(Ou dê duplo clique direto em `bin/moveops-tray.exe`)*.
3. 🚀 **Pronto!** Seu navegador de internet vai abrir automaticamente no endereço:
   ```
   http://localhost:8082
   ```

---

### 🕒 Como Usar o Ícone perto do Relógio (Bandeja do Sistema / System Tray)

Quando o MoveOps está aberto no Windows, ele fica minimizado de forma discreta perto do relógio (canto inferior direito da tela):

```
       +---------------------------------------------+
       | Abrir no Navegador (Dashboard)              |
       | ------------------------------------------- |
       | Status da Migração: Conectado               |
       | Pausar / Retomar Migração                   |
       | ------------------------------------------- |
       | Fechar e Sair do MoveOps                    |
       +---------------------------------------------+
                        [ Ícone MoveOps ] [ 15:30 ]
```

1. Procure pelo ícone do **MoveOps** ao lado do relógio do Windows (talvez seja necessário clicar na setinha para cima `^` para mostrar ícones ocultos).
2. **Clique com o Botão Direito** no ícone para acessar o menu rápido:
   - 🌐 **Abrir no Navegador:** Abre a tela do MoveOps imediatamente.
   - 📊 **Status:** Mostra se ele está transferindo arquivos ou ocioso.
   - ⏸️ **Pausar / Retomar:** Pausa temporariamente a cópia com um clique.
   - ❌ **Sair:** Fecha o MoveOps com segurança.

---

## 🐧 Como Instalar e Rodar no Linux (Ubuntu, Debian, Linux Mint e derivados)

Escolha o método mais prático para a sua distribuição:

### Método 1: Pacote `.deb` (Padrão e Recomendado no Linux)

Se você usa Ubuntu, Debian, Linux Mint ou Pop!_OS:

#### Pela Interface Gráfica (Sem Terminal):
1. Abra o gerenciador de arquivos e localize o pacote:
   ```
   dist/moveops_1.0.0_amd64.deb
   ```
2. **Dê duplo clique** no arquivo `.deb`.
3. A Central de Aplicativos (Ubuntu Software / Central de Programas) vai abrir.
4. Clique no botão verde ou azul **"Instalar"** e digite sua senha de usuário.
5. ✨ **Pronto!** O MoveOps agora aparecerá no seu **Menu de Aplicativos** (basta apertar a tecla `Super/Windows` e digitar `MoveOps`). Clique no ícone para abrir!

#### Pelo Terminal (Apenas 1 comando rápido):
```bash
sudo dpkg -i dist/moveops_1.0.0_amd64.deb
```

---

### Método 2: Modo Portátil (Roda Direto sem Instalação)
Se você não quer instalar pacotes e prefere rodar o binário direto da pasta:

1. Abra o terminal na pasta do projeto e execute:
   ```bash
   ./bin/moveops -open
   ```
2. A flag `-open` inicia o servidor e **já abre o navegador padrão automaticamente** em `http://localhost:8082`.

---

### ⚙️ Serviço em Segundo Plano no Linux (Opcional)
Se você instalou o MoveOps em um servidor Linux e quer que ele inicie sempre sozinho quando o computador ligar:

```bash
# Inicia o serviço
sudo systemctl start moveops

# Faz o MoveOps ligar automaticamente com o sistema
sudo systemctl enable moveops
```

---

## 🌐 Como Acessar o MoveOps a partir de Outro Computador na Rede

Se o MoveOps estiver rodando no computador da empresa ou em um servidor e você quiser controlá-lo a partir do seu notebook ou celular conectado na mesma rede Wi-Fi/cabo:

### 1. Descubra o IP do computador onde o MoveOps está rodando:
- **No Windows:** Abra o Prompt de Comando (CMD) e digite `ipconfig`. Procure pelo *Endereço IPv4* (exemplo: `192.168.1.50`).
- **No Linux:** Abra o terminal e digite `hostname -I`. O primeiro número é o seu IP (exemplo: `192.168.1.50`).

### 2. No outro computador:
Abra o navegador e digite o IP seguido de `:8082`:
```
http://192.168.1.50:8082
```

> [!TIP]
> **Dica de Ouro:** Não esqueça de colocar os dois pontos e o número da porta `:8082` no final do endereço.

---

## 🧭 Como Fazer a Sua Primeira Migração de Teste (Guia Rápido em 4 Passos)

Assim que abrir a tela do MoveOps no seu navegador:

1. **Aba "Armazenamento":** Veja os discos do seu computador e confira o espaço livre.
2. **Aba "Nova Migração":**
   - Escolha a **Origem** (pasta onde estão os arquivos que você quer copiar).
   - Escolha o **Destino** (pasta para onde os arquivos devem ir).
   - Mantenha selecionado o perfil **"Produção Padrão"** (já vem pronto com 50 MB/s e 300 IOPS, ideal para não atrapalhar o uso do computador).
3. **Clique em "Estimar / Pré-visualizar (Dry Run)":** Ele calcula quantos arquivos existem e quanto tempo vai demorar, sem copiar nada ainda.
4. **Clique em "Iniciar Migração":** A tela muda para a aba **"Telemetria"**, onde você vê o progresso em tempo real, velocidade e os arquivos sendo conferidos um a um com verificação de integridade matemática (`xxHash64`)!

---

## ❓ Perguntas Frequentes e Resolução Rápida de Problemas

### 1. "O navegador diz que 'A página não pode ser exibida' ou 'Conexão recusada'"
- **Causa provável:** O MoveOps ainda não foi iniciado ou foi fechado acidentalmente.
- **Solução:** Dê duplo clique em `Iniciar-MoveOps.bat` (no Windows) ou execute `./bin/moveops` (no Linux). Verifique se o ícone do MoveOps aparece na barra de tarefas/relógio.

### 2. "A porta 8082 já está sendo usada por outro programa"
- **Solução:** Você pode escolher qualquer outra porta livre (por exemplo 8090 ou 9000).
  - No Windows via terminal: `.\bin\moveops.exe -port 8090`
  - No Linux via terminal: `./bin/moveops -port 8090`
  - E acesse no navegador: `http://localhost:8090`.

### 3. "Como pausar a cópia se eu precisar de 100% da velocidade da internet agora?"
- Na aba **Telemetria** do navegador, basta clicar no botão **"Pausar"**. Quando terminar o que estiver fazendo, clique em **"Retomar"**. Ele continua exatamente de onde parou sem perder nada!
- Ou, no Windows, clique com o botão direito no ícone do relógio e escolha **"Pausar / Retomar"**.

### 4. "O MoveOps apaga meus arquivos de origem?"
- **NÃO!** O MoveOps apenas **lê** os arquivos da pasta de origem e grava cópias perfeitas na pasta de destino. Seus arquivos originais permanecem 100% intactos e intocados.

---

## 📞 Precisa de Ajuda ou Quer Detalhes Técnicos Avançados?

Se você é desenvolvedor, SysAdmin ou quer entender a fundo a matemática dos algoritmos e arquitetura de sistema:
- Consulte a documentação completa na pasta [`documentações/`](./documentações/00-INDICE-GERAL.md).
- Consulte o laudo de segurança em [`SECURITY_AUDIT.md`](./SECURITY_AUDIT.md).
- Consulte a especificação técnica em [`ARCHITECTURE.md`](./ARCHITECTURE.md).

---
*MoveOps — Enterprise Zero-Downtime Migration Engine*
