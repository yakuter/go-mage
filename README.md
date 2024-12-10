# go-mage

`go-mage`, Go projelerini derlemek ve yönetmek için `Makefile` yerine `magefile` kullanan bir projedir. `magefile`, Go dilinde yazılmış bir build aracıdır ve `Makefile`'a benzer şekilde çalışır, ancak Go'nun gücünü ve esnekliğini kullanır.

## Projenin Amacı

**go-mage** projesinin amacı, Go projelerini derlemek ve test etmek ve diğer build işlemlerini kolaylaştırmak için cross platform destekli örnek bir **magefile** sunmaktır.. `magefile`, Go dilinde yazıldığı için, Go geliştiricileri için daha tanıdık ve esnek bir yapı sunar.

## Projenin Özellikleri ve Uyarılar

- **Cross Platform Destek**: Proje **Windows**, **MacOS** ve **Linux** işletim sistemlerinde derlenebilir.

- **Windows Özel Özellikleri**: 
  - MinGW kullanarak Windows için derleme yapabilir
  - Syso ve version info dosyaları otomatik olarak üretilir
  - Windows'ta build almadan önce build/windows/assets klasörüne manifest (app.manifest) ve icon (app.icon) dosyalarını koymayı unutmayınız.

- **MacOS Özel Özellikleri**:
  - MacOS'ta istenilen SDK ile build alınabilir. Mage'i çalıştırmadan önce `export MACOS_SDK_VERSION=14.5` gibi env set etmeniz yeterli.

- **Güvenlik ve Kod Kalitesi**:
  - Vulnerability check ile güvenlik taraması yapar
  - Golangci-lint ile kod kalitesi kontrolü sağlar
  - Gocov ile test coverage raporlaması yapar
  - Magefile'da belirtilen versiyon ve diğer bilgiler kullanılarak projedeki Version, BuildTime, CommitID ve BuildMode değişkenleri build esnasında güncellenir.

## Ön Hazırlık

Projenin çalışabilmesi için sistemde `mage` uygulamasının kurulu olması gerekmektedir. `mage`'i kurmak için aşağıdaki adımları izleyebilirsiniz:
```sh
go install github.com/magefile/mage@latest
```

Kurulumdan sonra, mage komutunu kullanarak magefile'ı çalıştırabilirsiniz. Örneğin:
```sh
mage build
```

Alternatif olarak mage uygulamasını kurmadan magefile'ı doğrudan Go ile çalıştırabilirsiniz:
```sh
go run mage.go build
```

Son bir kullanım şekli olarak mage'i önceden build alabilir ve sunucuda öyle çalıştırabilirsiniz. Böylece sunucudaki bağımlılıklardan kurtulmuş olursunuz.
```sh
mage --compile ./runmage
./runmage
```

## Kurulum

Projeyi klonladıktan sonra gerekli bağımlılıkları yüklemek için aşağıdaki adımları izleyin:

```sh
git clone https://github.com/yakuter/go-mage.git
cd go-mage
go mod tidy
```

## Kullanım
Projede hangi mage komutları olduğunu görmek için `mage` komutunu çalıştırmanız yeterli.
```sh
➜ mage
Targets:
  build        builds the binary.
  clean        cleans the build directory.
  generate     runs go generate.
  linter       runs the linter.
  test         runs the tests.
  vulncheck    runs the vulnerability check.
```

Ardından istediğiniz bir komutu seçip direkt çalıştırabilirsiniz. Örneğin projedeki güvenlik zafiyetlerini Go'nun kendi aracıyla yakalamak için şu komutu çalıştırabilirsiniz:
```sh
mage vulncheck
```

## Katkıda Bulunma
Katkıda bulunmak isterseniz, lütfen bir pull request gönderin veya bir issue açın.

## Lisans
Bu proje MIT Lisansı ile lisanslanmıştır. Daha fazla bilgi için LICENSE dosyasına bakabilirsiniz.