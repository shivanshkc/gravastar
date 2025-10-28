# [1.1.0](https://github.com/shivanshkc/gravastar/compare/v1.0.2...v1.1.0) (2025-10-28)


### Bug Fixes

* **core:** gravity engine Close method ([73ae0af](https://github.com/shivanshkc/gravastar/commit/73ae0af483eed2ce4385a8ed09d89698462da245))
* **http:** add CORS ([a439494](https://github.com/shivanshkc/gravastar/commit/a4394940ac15172d33f3dadd8f2e8b5c2859681b))
* **http:** websocket write timeout bug fix ([095151f](https://github.com/shivanshkc/gravastar/commit/095151ff51d0e44f76bf2253bf9d8127357742f6))
* **physics:** better NaN protection and death sound ([827dfc5](https://github.com/shivanshkc/gravastar/commit/827dfc5ad08c5278a079d881bc224bb2fa5e5ce5))
* **physics:** defend against large deltas ([95e3816](https://github.com/shivanshkc/gravastar/commit/95e38161ad43bd132d70a59bc06ade6dc9afe1b5))
* **physics:** scale radius along with other vectors ([998429f](https://github.com/shivanshkc/gravastar/commit/998429f320216c7179a0d8e5624c5e9768ea63b6))
* **ui:** fix initial timer value, add TODO ([eee154e](https://github.com/shivanshkc/gravastar/commit/eee154e0d8d1f896888b226bf1e485d199009d08))
* **ui:** responsive instruction text ([1cbdf62](https://github.com/shivanshkc/gravastar/commit/1cbdf621157ac831b372c31dda9a78c81839894e))


### Features

* **physics:** do not allow overlapping dot creation ([672dcc9](https://github.com/shivanshkc/gravastar/commit/672dcc938e08b4f660db83a5c753e3d13599cb9f))
* **physics:** right wall collision removal ([37519a1](https://github.com/shivanshkc/gravastar/commit/37519a1efed8a8c713ca063881697a00206d59a9))
* **ui:** add collision sound effect ([efee40b](https://github.com/shivanshkc/gravastar/commit/efee40b011d8ed7af703b86b6be27bdaf62c116d))
* **ui:** add danger gradient for the right wall ([fd8dd0b](https://github.com/shivanshkc/gravastar/commit/fd8dd0b6a096e4a43fc7dfc873819ce28a3d1e31))
* **ui:** add dot trails ([20fcf7c](https://github.com/shivanshkc/gravastar/commit/20fcf7c75c6627eefc4cbcafcd2604f14e0d7ba9))
* **ui:** add instructional text with fade out ([fdab18c](https://github.com/shivanshkc/gravastar/commit/fdab18c10534694b32a1373c5b11d85e987db52c))
* **ui:** add timer ([decac2a](https://github.com/shivanshkc/gravastar/commit/decac2a7a8f147e6f77a4b0aa8cbb9a049b1aab8))
* **ui:** death sound and NaN protection ([23acdd7](https://github.com/shivanshkc/gravastar/commit/23acdd70ce5a34a3e392efeeb534bd89f7b4a5d2))
* **ui:** delete dot upon right wall collision ([079116a](https://github.com/shivanshkc/gravastar/commit/079116acebf680ba5a4a2f5025d5025d33bffda0))
* **ui:** hightlight boundaries for own dots ([411751d](https://github.com/shivanshkc/gravastar/commit/411751da8be057b8299a8f0e64786a5b86fad840))
* **ui:** mute button control ([73bb37e](https://github.com/shivanshkc/gravastar/commit/73bb37e0daeceb3307c673221f79889cf4fe5096))
* **ui:** theme-based dot coloring ([8121c01](https://github.com/shivanshkc/gravastar/commit/8121c019658f20b0a3979aa145a2cd88d724ecc4))
* **ui:** use bulma css, invert colors, move sync to bottom ([df4f53c](https://github.com/shivanshkc/gravastar/commit/df4f53cd631667cc6341e4aa1fa83b9948bf8b2c))

## [1.0.2](https://github.com/shivanshkc/gravastar/compare/v1.0.1...v1.0.2) (2025-10-17)


### Bug Fixes

* **ci:** update frontend config at runtime ([c2e80bc](https://github.com/shivanshkc/gravastar/commit/c2e80bc25a6f55f2c7b27e4fb83f9323d53d88b2))

## [1.0.1](https://github.com/shivanshkc/gravastar/compare/v1.0.0...v1.0.1) (2025-10-17)


### Bug Fixes

* **ci:** fix frontend configs upon deployment ([95bf8f7](https://github.com/shivanshkc/gravastar/commit/95bf8f78adb529cebb5f7e6d6a481f40e958dfed))

# 1.0.0 (2025-10-17)


### Bug Fixes

* **ci:** add ebiten dependencies, fix deploy action ([46c26df](https://github.com/shivanshkc/gravastar/commit/46c26df60d039ff259a6c36caed077cc2488d726))
* **ci:** include the static directory in the container image, and other docker fixes ([100258f](https://github.com/shivanshkc/gravastar/commit/100258f8b6e821cc3f8e394a93ce95c2a4a416d7))
* **docs:** add readme ([61f3480](https://github.com/shivanshkc/gravastar/commit/61f3480db64de544414a5d8423dc00d801ab98fe))
* **http:** create dot input validations ([8d487fc](https://github.com/shivanshkc/gravastar/commit/8d487fc6ac183e7a5ef892b7c783e35d6823c35d))
* **lint:** no linting for the ebiten file ([9156715](https://github.com/shivanshkc/gravastar/commit/91567151cf92c7a8cef27525e85ce4f51f24adeb))
* **physics:** run the engine, log request method in access logs ([f501982](https://github.com/shivanshkc/gravastar/commit/f50198228fab58332a26ee66e0372791fadbaf05))


### Features

* **http:** add create and list dots api ([b938c6d](https://github.com/shivanshkc/gravastar/commit/b938c6d72d3e1cf8f64aab158e1f9c58e2aea95a))
* **http:** remove Clear Dots API, add color input in Create Dots API ([9d0b648](https://github.com/shivanshkc/gravastar/commit/9d0b648eb27d778b61f78c4b692b409e78e11d9e))
* **http:** websocket integration ([a234057](https://github.com/shivanshkc/gravastar/commit/a23405739ab020381b257d0e2af5d002d4da614c))
* **physics:** add the gravity engine ([8694692](https://github.com/shivanshkc/gravastar/commit/869469297a84f0fe9468158a41f54454edb7f508))
* **physics:** target FPS based rendering ([3804b2d](https://github.com/shivanshkc/gravastar/commit/3804b2d64670213b0f56d219d2caa6507cd4b8ff))
* **ui:** add frontend ([30aa4f8](https://github.com/shivanshkc/gravastar/commit/30aa4f857a444577c859bfeaec40ffefe38ffac9))
