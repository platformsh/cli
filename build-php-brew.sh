set -ex

DIR=$1
PHP_VERSION=$2
OPENSSL_VERSION=$3

brew install bison pkg-config coreutils autoconf

SSL_DIR_PATH=$(pwd)/"$DIR"/ssl
mkdir -p "$SSL_DIR_PATH"

curl -LfSsl https://www.openssl.org/source/openssl-"$OPENSSL_VERSION".tar.gz | tar  xzf - -C "$DIR"
cd "$DIR"/openssl-"$OPENSSL_VERSION"

./config no-shared --prefix="$SSL_DIR_PATH" --openssldir="$SSL_DIR_PATH"
make
make install

cd ../..
curl -fSsl https://www.php.net/distributions/php-"$PHP_VERSION".tar.gz | tar  xzf - -C "$DIR"
cd "$DIR"/php-"$PHP_VERSION"

rm -f sapi/cli/php

./buildconf --force
./configure \
  --disable-shared \
  --enable-embed=static \
  --enable-filter \
  --enable-pcntl \
  --enable-phar \
  --enable-posix \
  --enable-static \
  --enable-sysvmsg \
  --with-curl \
  --with-openssl \
  --with-pear=no \
  --without-pcre-jit \
  --with-zlib \
  --disable-all \
OPENSSL_CFLAGS="-I$SSL_DIR_PATH/include" \
OPENSSL_LIBS="-L$SSL_DIR_PATH/lib -lssl -lcrypto"

make -j"$(nproc)" cli
