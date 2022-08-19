set -ex

BREW_PREFIX=$(brew --prefix)
export PKG_CONFIG_PATH="$PKG_CONFIG_PATH:$BREW_PREFIX/opt/openssl@1.1/lib/pkgconfig"

DIR=$2
mkdir -p $DIR
curl -fSsl https://www.php.net/distributions/php-$1.tar.gz | tar  xzf - -C $DIR
cd $DIR/php-$1

rm -f sapi/cli/php

./buildconf --force
./configure \
  --disable-shared \
  --enable-embed=static \
  --enable-mbstring \
  --enable-pcntl \
  --enable-phar \
  --enable-posix \
  --enable-static \
  --enable-sysvmsg \
  --with-openssl \
  --with-pear=no \
  --without-pcre-jit \
  --disable-all

export PATH="$BREW_PREFIX/opt/bison/bin:$PATH"
make -j$(sysctl -n hw.logicalcpu) cli
