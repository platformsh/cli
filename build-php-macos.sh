set -ex

curl -fSsl https://www.php.net/distributions/php-$1.tar.gz | tar xzf -
cd php-$1
rm -f sapi/cli/php

export PKG_CONFIG_PATH="$PKG_CONFIG_PATH:/opt/homebrew/opt/openssl@1.1/lib/pkgconfig"
./buildconf --force
./configure \
  --enable-embed=static \
  --enable-mbstring \
  --enable-phar \
  --enable-static=yes \
  --enable-sysvmsg \
  --with-openssl \
  --enable-pcntl \
  --enable-posix \
  --with-pcre-jit \
  --disable-all

export PATH="/opt/homebrew/opt/bison/bin:$PATH"
make -j$(sysctl -n hw.logicalcpu) cli
