axe
========

![axe](https://media.giphy.com/media/l17uofKSRXJGIsnYNB/giphy.gif)

## Demo

Rio
[![asciicast](https://asciinema.org/a/pQOUSFp3S3uANAMMrsEKiuo1e.svg)](https://asciinema.org/a/pQOUSFp3S3uANAMMrsEKiuo1e)

K8s
[![asciicast](https://asciinema.org/a/KX2pvUZjtNDjEzGJs6LBDEiGX.svg)](https://asciinema.org/a/KX2pvUZjtNDjEzGJs6LBDEiGX)

## Building

`make`

## Running

`./bin/axe --kubeconfig $KUBECONFIG`

## Example

1. Define your root page, shortcuts, viewmap, pageNav, footers, tableEventHandler
2. Throwing!

```$xslt

drawer = types.Drawer{
	RootPage:  RootPage,
	Shortcuts: Shortcuts,
	ViewMap:   ViewMap,
	PageNav:   PageNav,
	Footers:   Footers,
}
	
app := throwing.NewAppView(clientset, drawer, tableEventHandler)
if err := app.Init(); err != nil {
	return err
}

return app.Run()

```

## Contribution

https://github.com/derailed/k9s

https://github.com/rivo/tview

https://github.com/gdamore/tcell

## License
Copyright (c) 2019 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
