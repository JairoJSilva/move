Estou com a ideia de fazer uma aplicação para migrar arquivos de um disco para o outro de forma granular, gradativa e controlada para que não afete a produção do cliente.

Essas migrações devem ser feita de forma que não precise que o cliente para de ler e/ou escrever e nem precise parar fazer isso,mesmo que seja de volumetria muito grande. 100Tb por exemplo.

Essa aplicação deve rodar tanto em Windows como em Linux ou de Linux para windows e vice e versa.
Utilize a linguagem mais performartica para isso e que seja compativel com ambos os sistemas operacionais.


Porem, ela deve ter um frontend que seja bem interativo para que qualquer pessoa consiga realizar a migração.
Ela também tem a obrigatoriedade de conseguir ler e reconhecer os discos e volumes que existem no host para facilitar os apontamentos.
Ela também deve conter metricas parametrizaveis para a migração.

Seja, sincrinozar tudo com outro disco, migrar por tada,dia, mês e/ou ano.
Realizar sync de apenas os arquivos que estão diferentes emtre duas pastas, discos e/ou diretorio.
E escolher a opção de se quer ou não que crie um relatorio detalhado e sumarizado de tudo que foi copiado. 
Para que seja possivel auditar posteriormente.
