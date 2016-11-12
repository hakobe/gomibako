/* eslint-disable import/no-extraneous-dependencies */
import gulp from 'gulp';
// import babel from 'gulp-babel';
import eslint from 'gulp-eslint';
import del from 'del';
import webpack from 'webpack-stream';
import webpackConfig from './webpack.config.babel';

const paths = {
  allSrcJs: 'js/src/**/*.js',
  gulpFile: 'gulpfile.babel.js',
  webpackFile: 'webpack.config.babel.js',
  entryPoint: 'js/src/app.js',
  distDir: 'static/script',
};

gulp.task('clean', () => del(paths.libDir));

gulp.task('build', ['lint', 'clean'], () =>
  gulp.src(paths.allSrcJs)
    .pipe(webpack(webpackConfig))
    .pipe(gulp.dest(paths.distDir))
);

gulp.task('default', ['build']);

gulp.task('lint', () =>
  gulp.src([
    paths.allSrcJs,
    paths.gulpFile,
  ])
    .pipe(eslint())
    .pipe(eslint.format())
    .pipe(eslint.failAfterError())
);
