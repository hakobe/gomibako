/* eslint-disable import/no-extraneous-dependencies */
import gulp from 'gulp';
// import babel from 'gulp-babel';
import eslint from 'gulp-eslint';
import postcss from 'gulp-postcss';
import autoprefixer from 'autoprefixer';
import postcssnested from 'postcss-nested';
import del from 'del';
import webpack from 'webpack-stream';
import webpackConfig from './webpack.config.babel';

const paths = {
  allSrcJs: 'js/**/*.js?(x)',
  allSrcCss: 'css/**/*.css',
  gulpFile: 'gulpfile.babel.js',
  webpackFile: 'webpack.config.babel.js',
  entryPointJs: 'js/reqevents.jsx',
  distDirJs: 'static/script/',
  distDirCss: 'static/style/',
};

gulp.task('clean', () => del([paths.distDirJs, paths.distDirCss]));

gulp.task('build-js', ['lint', 'clean'], () =>
  gulp.src(paths.entryPointJs)
    .pipe(webpack(webpackConfig))
    .pipe(gulp.dest(paths.distDirJs))
);

gulp.task('build-css', ['lint', 'clean'], () =>
  gulp.src(paths.allSrcCss)
    .pipe(postcss([
      postcssnested,
      autoprefixer({
        browsers: ['last 1 version'],
      }),
    ]))
    .pipe(gulp.dest(paths.distDirCss))
);

gulp.task('build', ['build-js', 'build-css']);

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

gulp.task('watch', () => {
  gulp.watch([paths.allSrcJs, paths.allSrcCss], ['default']);
});
