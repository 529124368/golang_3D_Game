package shader

var Shader []byte

func init() {
	Shader = []byte(`
	package main
	func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
		pos := texCoord
		offset :=1/imageSrcTextureSize()
		if  imageSrc0UnsafeAt(vec2(pos.x,pos.y)).a==0 && (imageSrc0UnsafeAt(vec2(pos.x+offset.x,pos.y)).a!=0 || imageSrc0UnsafeAt(vec2(pos.x-offset.x,pos.y)).a!=0 || imageSrc0UnsafeAt(vec2(pos.x,pos.y+offset.y)).a!=0 || imageSrc0UnsafeAt(vec2(pos.x+offset.x,pos.y-offset.y)).a!=0){
			return vec4(100,100,0,1)
		}
		return imageSrc0UnsafeAt(texCoord)
	}
`)
}
