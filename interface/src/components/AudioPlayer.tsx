type AudioProps = {
  src: string
}
const AudioPlayer = ({ src }: AudioProps) => {
  return (
    <audio controls>
      <source src={src} type="audio/mpeg" style={{ width: '100%' }} />
      Your browser does not support the audio element.
    </audio>
  )
}

export default AudioPlayer
