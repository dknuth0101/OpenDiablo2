package d2systems

import (
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2interface"
	"github.com/gravestench/akara"
	"path/filepath"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2components"
)

const (
	mainViewport int = 0
)

func NewBaseScene(key string) *BaseScene {
	base := &BaseScene{
		BaseSystem: &akara.BaseSystem{},
		Logger: d2util.NewLogger(),
		key: key,
		Viewports: make([]akara.EID, 0),
		GameObjects: make([]akara.EID, 0),
		systems: &baseSystems{},
	}

	base.SetPrefix(key)

	return base
}

var _ akara.System = &BaseScene{}

type baseSystems struct {
	*RenderSystem
	*GameObjectFactory
}

type BaseScene struct {
	*akara.BaseSystem
	*d2util.Logger
	key string
	booted bool
	paused bool
	systems *baseSystems
	Add *sceneObjectAssigner
	Viewports []akara.EID
	GameObjects []akara.EID
	*d2components.MainViewportMap
	*d2components.ViewportMap
	*d2components.ViewportFilterMap
	*d2components.CameraMap
	*d2components.SurfaceMap
	*d2components.PositionMap
	*d2components.ScaleMap
	*d2components.AnimationMap
}

func (s *BaseScene) Booted() bool {
	return s.booted
}

func (s *BaseScene) Paused() bool {
	return s.paused
}

func (s *BaseScene) Init(world *akara.World) {
	s.World = world

	if s.World == nil {
		return
	}
}

func (s *BaseScene) boot() {
	s.Info("booting ...")

	s.injectComponentMaps()

	s.Add = &sceneObjectAssigner{BaseScene: s}

	for idx := range s.Systems {
		if rendersys, ok := s.Systems[idx].(*RenderSystem); ok {
			s.systems.RenderSystem = rendersys
			continue
		}

		if objFactory, ok := s.Systems[idx].(*GameObjectFactory); ok {
			s.systems.GameObjectFactory = objFactory
			continue
		}
	}

	if s.systems.RenderSystem == nil {
		s.Info("waiting for render system ...")
		return
	}

	if s.systems.RenderSystem.renderer == nil {
		s.Info("waiting for renderer instance ...")
		return
	}

	if s.systems.GameObjectFactory == nil {
		s.Info("waiting for game object factory ...")
		return
	}

	s.systems.SpriteFactory.RenderSystem = s.systems.RenderSystem

	s.createDefaultViewport()

	s.Info("booted!")
	s.booted = true
}

func (s *BaseScene) injectComponentMaps() {
	s.MainViewportMap = s.World.InjectMap(d2components.MainViewport).(*d2components.MainViewportMap)
	s.ViewportMap = s.World.InjectMap(d2components.Viewport).(*d2components.ViewportMap)
	s.ViewportFilterMap = s.World.InjectMap(d2components.ViewportFilter).(*d2components.ViewportFilterMap)
	s.CameraMap = s.World.InjectMap(d2components.Camera).(*d2components.CameraMap)
	s.SurfaceMap = s.World.InjectMap(d2components.Surface).(*d2components.SurfaceMap)
	s.PositionMap = s.World.InjectMap(d2components.Position).(*d2components.PositionMap)
	s.ScaleMap = s.World.InjectMap(d2components.Scale).(*d2components.ScaleMap)
	s.AnimationMap = s.World.InjectMap(d2components.Animation).(*d2components.AnimationMap)
}

func (s *BaseScene) createDefaultViewport() {
	viewportID := s.NewEntity()
	s.AddViewport(viewportID)

	camera := s.AddCamera(viewportID)
	camera.Width = 800
	camera.Height = 600
	camera.Zoom = 1

	s.AddSurface(viewportID)
	s.AddMainViewport(viewportID)

	s.Viewports = append(s.Viewports, viewportID)
}

func (s *BaseScene) Key() string {
	return s.key
}

func (s *BaseScene) Update() {
	if !s.booted {
		s.boot()
	}

	if !s.booted {
		return
	}

	s.RenderViewports()
}

func (s *BaseScene) RenderViewports() {
	if s.systems.RenderSystem == nil {
		return
	}

	if s.systems.RenderSystem.renderer == nil {
		return
	}

	numViewports := len(s.Viewports)

	if numViewports < 1 {
		return
	}

	viewportObjects := s.binGameObjectsByViewport()

	for idx := numViewports-1; idx >= 0; idx-- {
		s.renderViewport(idx, viewportObjects[idx])
	}
}

func (s *BaseScene) binGameObjectsByViewport() map[int][]akara.EID {
	bins := make(map[int][]akara.EID)

	for _, eid := range s.GameObjects {
		vpfilter, found := s.GetViewportFilter(eid)
		if !found {
			vpfilter = s.AddViewportFilter(eid)
			vpfilter.Set(mainViewport, true)
		}

		for _, vpidx64 := range vpfilter.ToIntArray() {
			vpidx := int(vpidx64)

			_, found := bins[vpidx]
			if !found {
				bins[vpidx] = make([]akara.EID, 0)
			}

			bins[vpidx] = append(bins[vpidx], eid)
		}
	}

	return bins
}

func (s *BaseScene) renderViewport(idx int, objects []akara.EID) {
	id := s.Viewports[idx]

	if idx == mainViewport {
		s.AddMainViewport(id)
	} else {
		s.MainViewportMap.Remove(id)
	}

	camera, found := s.GetCamera(id)
	if !found {
		return
	}

	sfc, found := s.GetSurface(id)
	if !found {
		return
	}

	if sfc.Surface == nil {
		sfc.Surface = s.systems.renderer.NewSurface(camera.Width, camera.Height)
	}

	cx, cy := int(camera.X()) + camera.Width>>1, int(camera.Y()) + camera.Height>>1

	sfc.Surface.PushTranslation(-cx, -cy) // negative because we're offsetting everything that gets rendered
	sfc.Surface.PushScale(-camera.Zoom, -camera.Zoom)

	for _, object := range objects {
		s.renderObject(sfc.Surface, object)
	}

	sfc.Pop()
	sfc.Pop()
}

func (s *BaseScene) renderObject(target d2interface.Surface, id akara.EID) {
	sfc, found := s.GetSurface(id)
	if !found {
		return
	}

	position, found := s.GetPosition(id)
	if !found {
		position = s.AddPosition(id)
	}

	scale, found := s.GetScale(id)
	if !found {
		scale = s.AddScale(id)
	}

	sfc.PushTranslation(int(position.X()), int(position.Y()))
	sfc.PushScale(scale.X(), scale.Y())

	target.Render(sfc.Surface)

	sfc.Pop()
	sfc.Pop()
}

// responsible for wrapping the object factory calls and assigning the created object entity id's to the scene
type sceneObjectAssigner struct {
	*BaseScene
}

func (s *sceneObjectAssigner) Sprite(x, y float64, imgPath, palPath string) akara.EID {
	s.Infof("creating sprite: %s, %s", filepath.Base(imgPath), palPath)

	eid := s.systems.SpriteFactory.Sprite(x, y, imgPath, palPath)
	s.GameObjects = append(s.GameObjects, eid)

	return eid
}
