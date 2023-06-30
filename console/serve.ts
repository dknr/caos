import { withArgs, withDefaults } from "../cmd.ts";
import { serveCaos } from "../server/mod.ts";
import { openCaos } from "../store/mod.ts";

export default withArgs(withDefaults({
  home: 'd10b49b4cf4f9204c4a6e4a96e5a004fa25768623667b2aec05f82e4852aaa91',
  path: '/home/dknr/caos',
}, ({home, path}) => {
  const caos = openCaos({path});
  serveCaos(caos, {home});
}));
