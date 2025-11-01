# slide 1
Dobrý den, jmenuju se Lukáš Stuchlík a rád bych vám tady něco řekl o alternativě k Helmu, aneb jak se nezbláznit z YAML templatů

# slide 2
V kostce o mně, vystudoval jsem magisterské studium na Ostravské univerzitě v oboru Informační systémy, a od té doby dělám už 5 let pro ostravskou firmu ProRocketeers. Jako firma děláme agilní vývoj softwaru pro klienty, kde převážně nabízíme Team as a Service. Já osobně se věnuju částečně nějakému backend vývoji v Typescriptu a Go, ale převážně jsem jako DevOps Engineer - správa CI/CD pipelines, Kubernetes clusterů, dělám s technologiemi jako Helm, Docker, Ansible, Terraform apod. Ve volném čase se rád věnuju tomu přispět něco do opensource komunity zpátky, mimo jiné i do projektu, který vám tady budu představovat

# slide 3
Helm charty jsou známý způsob, jakým se dají popsat libovolné Kubernetes aplikace pomocí YAML šablon. Jenže když do toho začnete míchat *příliš* moc logiiky, může z toho vzniknout nečitelná změť jako třeba tohle. Rád bych vám ukázal, že to jde i jinak

# slide 4
Prakticky stejný kus kódu, ale už se to dá číst i po obědě a necháte ho v sobě..

# slide 5
Jak jsme se k tomuhle dostali? U nás v ProRocketeers míváme i různé malé služby a projekty, které potřebují taky nasadit na některé z našich vlastních clusterů. Dříve jsme to řešili typicky tak, že jsme používali předgenerované Helm charty, které jsme si upravovali dle potřeb daného projektu. Tento přístup ale byl strašně repetitivní, a když jsme se rozhodli dělat nějaké změny globálně napříč všemi aplikacemi, dalo to vždycky hroznou spoustu práce to upravit všude. Vzal jsem si tedy na sebe nelehký úkol, a dal se do přípravy našeho "obecného" Helm chartu, který bude centralizovaný a společný pro všechny takovéhle projekty. Výhodou takového řešení je, že změny se provádějí už na jednom místě, drasticky to zkrátí nasazení nového projektu, a jednoduše se to verzuje.

Vzhledem k tomu, že jsem k tomu chtěl mít jaksi.. hezký formát těch vstupních dat, aby i developer experience (tedy naše) s používáním tohoto chartu byla přívětivá, nevyhnutelně přišla na řadu dost velké množství logiky. Různý preprocessing těchto dat, validace, úprava do formátu ve kterém se těm šablonám s tím lépe pracuje atd. A přesně tady začínají pramenit problémy s Helm charty..

# slide 6
Podpora IDE pro Helm templaty je dost mizerná, sotva se mi podařilo najít nějaký extension pro syntax highlighting, ale na takové základní věci jako IntelliSense pro napovídání datové struktury těch objektů jsem už mohl zapomenout a místo toho to držet v hlavě nebo někde bokem. S tímto ještě půl bídy, chybové hlášky v případě že se snažíte dereferencovat proměnnou, která třeba neexistuje jsou ještě vcelku ok, tam vám Helm alespoň řekne nějaké vodítka, třeba název toho co se snažíte přečíst..

Horší problém pramení z toho, že těmi šablonami vytváříte YAML soubory, a ty jsou závislé na odsazení - whitespace. Helm templaty, resp. ty dynamické části, které píšete do dvou složených závorek, totiž mají na každé straně optional pomlčku, a ta tomu enginu říká, že z téhle strany "spolkni" veškerý whitespace až do prvního normálního znaku. A řeknu vám, když koukáte do stovek řádků takovýchto template, tak je *strašně jednoduché* někde takovou pomlčku buď zapomenout, anebo ji dát někam, kde být nemá. A nikdo vás na to neupozorní, že je něco špatně až na to, že vám začne render padat na YAML syntax chybách, ze kterých nejde poznat, kde je co špatně. Nezbývá vám pak nezbývá nic moc jiného než pozorně pročítat řádek po řádku a hledat, kde je zdroj takové chyby. 

Další problém - řekněme, že si chcete ušetřit budoucí nervy, a rozhodnete se zavést nějaké testování, ideálně unit testování do svého Chartu. Ne že by to nešlo, existuje na to Helm plugin - `helm-unittest`, kdy i ty samotné unit testy píšete v YAMLech, ale ani práce s ním není úplně bezchybná (např. jsem se dostával do nekonzistencí, že lokálně při vývoji mi testy procházely, ale v CI/CD pipeline už ne a nepřišel jsem na to proč). 

Dlouho mi tyto problémy ležely v žaludku a chtěl jsem něco, čím bych to nahradil, a to mě přivádí k naší alternativě, kterou vám chci představit - Yoke.

# slide 7

Yoke je package manažer stejně jako Helm, ale místo křehkých YAML šablon se vydal programátorskou cestou, kde Kube objekty (nebo resourcy) vytváříte v programovacím jazyce, například Go. 

Základní jednotkou v Yoku je ekvivalent Helm Chartu, a nazývá se Flight. Yoke celkově bere svoje názvosloví z aviatiky, ale dá se na to zvyknout. Flight je jakýkoliv program, který je zkompilovaný do WebAssembly formátu, a který na vstupu bere nějaké parametry, a na výstupu dává JSON pole Kube resourců. Proč zrovna WebAssembly? Z několika důvodů: 

1) Je to společný kompilační target pro více různých jazyků
2) Je nezávislý od cílového OS nebo architektury
3) Jeho runtime je sandboxed, tj. nemá přístup k souborovému systému anebo síti

# slide 8
I když můžete teoreticky psát Flighty v různých jazycích, je vcelku doporučované vybrat si právě Go, a to z pár důvodů - jednak Yoke distribuuje sadu pomocných SDK právě pro Go, ale hlavní důvod za mě je oficiální podpora Kubernetes API typů, které si importujete do vašeho Flightu jako normální balíčky. Tím rovnou narážím na (z mého pohledu) hlavní výhodu, a to je plná podpora programovacího jazyka vaší volby. Můžete si svůj Flight napsat jako klasický Go program, používat všechny featury tohoto jazyka (i to co vám Helm v templatách nedovolil), máte k tomu plnou podporu IDE, klasickou type safety (a to nejen ve vašich datových strukturách ale právě i v Kube API strukturách), unit testing, dokonce i klasický breakpoint debugging. 

Pro usnadnění případného přechodu z Helmu si dokonce zachovává jistou zpětnou kompatibilitu s Helmem, a to v té formě, že můžete z Go kódu vyrenderovat už existující Helm chart (například nějaké závislosti, které nemůžete přepsat taky).

Dalo to nějakou práci ten náš Chart přepsat do Yoke implementace, ale ta jistota, kterou jsem tímto refactorem nabyl, se mi zatím zdá dost k nezaplacení. 

# slide 9 + 10
Připravil jsem si pro vás tady krátkou ukázku toho, jak implementace takového Flightu může vypadat. V prvé řadě si zadefinujeme strukturu našich vstupních dat. Řekněme, že půjde o jednoduchou aplikaci, kde si nechám vytvořit pouze jednoduchý Deployment a k ní Service. Bude mě teda zajímat její název, namespace, image toho kontejneru, a pak třeba počet replik a nějaký port na kterém bude naslouchat.

Ze standardního vstupu si pak načtu data a zparsuju je z JSONU nebo YAMLu do téhle struktury. V tomto bodě si vlastně můžu s těmito daty dělat *co chci*, na to Yoke neklade žádné omezení. Pro naše demo si tady tedy vytvořím jenom 2 malé funkce, které mi vezmou ty moje vstupní data, a vytvoří z nich ty Kubernetí resourcy. Tyhle resourcy pak jen posbírám, serializuju do JSON pole a zapíšu na standardní výstup.

Ty jednotlivé funkce pak můžou vypadat třeba takto. Kód jsem trochu zkrátil ať se mi to vejde na slide, ale jsem si jistý že pointu pochopíte. Tady můžete taky vidět, že vracím přímo objekty z zmiňovaného Kubernetes API, takže se nemusím obávat o to, že někde udělám nějaký typo ve stylu "replicas" versus "replica". Do objektu se vyplní nějaké povinné metadata jako právě třeba název a namespace, a pak už vyplňujete klasicky Deployment specku. Všechno je staticky typované, takže prakticky můžete tabulátorem vyplňovat všechny fieldy co potřebujete, Go vám napoví.

Service se potom vytváří naprosto ekvivalentním způsobem, tak to tady pro stručnost už vynechám. No... A to je vlastně všechno! Jak takovýhle Flight vlastně můžete vyzkoušet?

# slide 11 
Pořád je to program v Go, takže to můžete klasicky spustit (jq pro přehlednost výstupu). Pokud máte něco složitějšího, doporučuju vám k tomu ručně přiimplementovat možnost čtení ze souboru, lépe si s tím rozumí třeba VSCode debugger.

Nebo to můžete zkompilovat právě do WebAssembly, a pak spustit. K tomu se využívá subcommand Yoke CLI `yoke takeoff`, což je takový ekvivalent `helm install`. Nezapomeňte na flag `-dry`, ať si to omylem rovnou nenasadíte.

Jakmile máte jistotu, že to funguje jak má, nezbývá než to někde uploadnout nebo publishnout. Yoke si umí poradit s Flighty na klasickém HTTPS file serveru, S3 storage, ale zvládá taky třeba OCI artefakty jako třeba GitHub Container Repository, Azure Container Repository, Harbor atd. K takovému účelu pak poslouží command `yoke stow`, který ten Flight zabalí, zkomprimuje a uploadne.

# slide 12 + 13
Ve dnešní době je zvykem jít ve správě Kube clusterů směrem GitOps, kde v hlavní roli figuruje ArgoCD. Jak si Yoke poradí proti Helmu tady?

Helm je v Argu integrovaný nativně, vysloveně nepotřebuje žádné nastavení, prostě out of the box. Příklad nasazení Helm chartu s Argem vidíte na slidu. Nutno podotknout, že Argo používá Helm pouze pro templating, na samotné nasazení ho nepoužívá.

Oproti tomu Yoke je implementovaný jako tzv. Content Management Plugin. Argo vám tímto dává možnost integrovat do něj různé jiné způsoby vytváření Kube objektů, a realizuje se to tím, že se upraví jeho `repo-server` pod, kde se přidá další plugin kontejner. Bohužel, použití takového pluginu už je trošku ukecanější, ale v zásadě je to podobné tomu, co vyplňujete i v případě Helmu.

# slide 14
Yoke má ještě jednu zajímavost navíc oproti Helmu, kterou bych chtěl aspoň v rychlosti zmínit. Je to trochu individuální a někomu to může být jedno, ale Helm se potýká s nedostatečnou validací vstupů. Ano, je tu ta možnost sepsat validační JSON schéma pro vstupní data v Chartu, ale ty doposud nejsou povinné. V opačném případě nezbývá než se spoléhat na nějakou "aplikační" logiku, anebo prostě jen doufat že to funguje jak má.

Yoke se snaží tento problém řešit konceptem, který nazval ATC - Air Traffic Controller. High level jde o možnost si vcelku jednoduše vytvářet vlastní Custom Resourcy, které reprezentují tu vaši aplikaci, a ta je pak *implementovaná* Yoke Flightem, který si napíšete. Taková Custom Resource má pak *povinně* definované OpenAPI schéma, které se vám už o validaci těch vstupních dat postará. V opačném případě vás totiž zablokuje vysloveně Kube API server formou admission webhooku, a vysloveně vám nedovolí nasadit něco nevalidního.

Funguje to tak, že vy si nejprve naimplementujete Flight tak jako dříve.. Ale potom k němu sepíšete instanci objektu `Airway`. Tahle custom resource od Yoku vlastně definuje tu *vaši* Custom Resource, a zároveň k ní připojuje cestu k tomu implementujícímu Flightu. Když takový objekt nasadíte do clusteru, je načtený Air Traffic Controllerem (což je klasický Deployment), a na základě té Airway vám vytvoří klasickou Custom Resource Definition té vaší resourcy, a zároveň automaticky vytvoří controller pro tu resource, který ty objekty umí zpracovávat.

Vám potom už stačí do clusteru nasadit vysloveně ten váš custom typ, a ATC ho zpracuje a nasadí vám přímo ty objekty do clusteru.

Podle mě je tohle poměrně zajímavá feature, protože z vlastní zkušenosti vím, že psát něco takového "postaru" s nástroji jako Kubebuilder, Controller-manager nebo Operator SDK je *daleko* větší práce a tohle umí dosáhnout prakticky stejného výsledku daleko jednodušeji.

# slide 15
Blížíme se ke konci, takže srovnejme si oba tooly, nějaké pro a proti. Nejprve Helm.
- ruku na srdce, je battle tested - v provozu je asi 9-10 let, od roku 2018 je pod taktovkou CNCF (Cloud Native Computing Foundation) a plně akceptovaný jako "Graduated" projekt byl v roce 2020. Používají ho stovky firem a tisíce uživatelů po celém světě
- ve většině případů se poměrně jednoduše používá
- má nativní support v ArgoCD i jiných systémech, např. FluxCD
- má možnost používat sub-charty pro nasazení složitějších aplikací
- můžete použít i tzv. Helm hooky pro implementaci složitějších nasazovacích mechanismů/akcí

Na druhou stranu:
- dá se uvažovat, že není úplně bezpečný, z hlediska že není tak jednoduché nějak zkontrolovat co přesně si nasazujete, nebo můžou být charty podvrženy
- v případě složitých chartů, co obsahují spoustu logiky, tak ta udržovatelnost jde hodně ke dnu, a absence nějakých vestavěných testovacích mechanismů to jen zdůrazňuje

# slide 16
Yoke oproti tomu:
- využívá klasických programovacích jazyků ke tvorbě Kube resourců, namísto textových YAML template
- díky tomu má plnou podporu IDE, typovou bezpečnost, testování atd.
- má vcelku slušnou zpětnou kompatibilitu s Helmem umožňující plynulejší přechod
- má možnost aplikace reprezentovat díky ATC ve formě custom resourců

Na druhou stranu:
- částečně trpí podobnými problémy jako Helm - pořád si můžete deploynout něco jiného než si myslíte
  - na druhou stranu, veřejně dostupné Flighty jsou prakticky neexistující, což se rozhodně nedá říct o Helm chartech
- jako osobně největší potenciální mínus vidím to, že je to velice nový projekt (první commit v lednu 2024), který nemá žádné sponzory, a prakticky jednoho maintainera
  - pro některé toto může být důvod to jako možnost zamítnout, ale osobně si myslím, že i v tom stavu v jakém to je teď, je to vice než schopné a stačilo by udělat třeba fork nebo kopie docker imagí apod., pokud byste měli strach o to, že by to autor zabalil

# slide 17
Tohle je z mé strany asi vše, děkuji vám všem za pozornost a nechávám vám prostor na případné otázky
...
Pokud byste měli další otázky, které už jsme nestihli probrat, najdete mě ještě chvilku tady v prostorách před místností, v opačném případě neváhejte se mi ozvat na mail `lukas.stuchlik@prorocketeers.com`.
