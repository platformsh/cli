set -ex

brew install bison openssl@1.1 oniguruma pkg-config coreutils autoconf
BREW_PREFIX=$(brew --prefix)
export PKG_CONFIG_PATH="$PKG_CONFIG_PATH:$BREW_PREFIX/opt/openssl@1.1/lib/pkgconfig:$BREW_PREFIX/opt/oniguruma/lib/pkgconfig"

DIR=$2
mkdir -p $DIR
curl -fSsl https://www.php.net/distributions/php-$1.tar.gz | tar  xzf - -C $DIR
cd $DIR/php-$1

rm -f sapi/cli/php

./buildconf --force
./configure \
  --disable-shared \
  --enable-embed=static \
  --enable-filter \
  --enable-mbstring \
  --enable-pcntl \
  --enable-phar \
  --enable-posix \
  --enable-static \
  --enable-sysvmsg \
  --with-curl \
  --with-openssl \
  --with-pear=no \
  --without-pcre-jit \
  --disable-all

make -j$(nproc) cli
